package repo

import (
	"errors"
	"fmt"
	"pr-mail/app/domain"
	"pr-mail/app/dto"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PrRepo interface {
	GetUserByUsername(username string) (*domain.Credential, error)
	UpdateUserToken(userID uint, token string) error
	GetEmployeeByEmpID(empID string) (*domain.Employee, error)
	GetPRByEmpIDAndLink(employeeID uint, prLink string) (*domain.PullRequest, error)
	UpdatePullRequest(pr *domain.PullRequest) error
	SavePullRequest(pr *domain.PullRequest) error
	ValidPRByEmpID(empID string) (*domain.PullRequest, error)
	UpdatePRDetails(prID int64, data *dto.GitHubPRResponse) error
	SavePRSnapshot(snapshot *domain.PRSnapshot) error
	GetLatestSnapshot(prID uint) (*domain.PRSnapshot, error)
	BuildReportFromSnapshot(pr *domain.PullRequest, snap *domain.PRSnapshot) string
	SavePRReport(report *domain.PRReport) error
	FetchReportsByStaffID(empIDs []string) ([]domain.PRReport, error)
	BuildMailBody(reports []domain.PRReport) string
	GetAllOpenOrDraftPRsByEmpID(empID string) ([]domain.PullRequest, error)
	GetTodaySnapshotsByEmpID(empID string) ([]domain.PRSnapshot, error)
	GetPRByID(prID uint) (*domain.PullRequest, error)
	ReportExistsForEmpAndDate(empID string, date time.Time) (bool, error)
	ReportExistsForEmpAndDateAndPR(empID string, date time.Time, prLink string) (bool, error)
	MarkReportsAsMailed(reportIDs []uint) error
}

type PrRepoImpl struct {
	db *gorm.DB
}

func NewPrRepo(db *gorm.DB) PrRepo {
	return &PrRepoImpl{
		db: db,
	}
}

func (r *PrRepoImpl) GetUserByUsername(username string) (*domain.Credential, error) {
	var userDet domain.Credential
	if err := r.db.Table("credentials").Where("username = ?", username).First(&userDet).Error; err != nil {
		return nil, err
	}
	return &userDet, nil
}

func (r *PrRepoImpl) UpdateUserToken(userID uint, token string) error {
	return r.db.Table("credentials").
		Where("id = ?", userID).
		Update("token", token).Error
}

func (r *PrRepoImpl) GetEmployeeByEmpID(empID string) (*domain.Employee, error) {
	var emp domain.Employee
	err := r.db.Table("employees").Where("emp_id = ?", empID).First(&emp).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *PrRepoImpl) GetPRByEmpIDAndLink(employeeID uint, prLink string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	err := r.db.Table("pull_requests").
		Where("employee_id = ? AND pr_link = ?", employeeID, prLink).
		First(&pr).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &pr, nil
}

func (r *PrRepoImpl) UpdatePullRequest(pr *domain.PullRequest) error {
	return r.db.Table("pull_requests").Save(pr).Error
}

func (r *PrRepoImpl) SavePullRequest(pr *domain.PullRequest) error {
	return r.db.Table("pull_requests").Save(pr).Error
}

func (r *PrRepoImpl) ValidPRByEmpID(empID string) (*domain.PullRequest, error) {
	var employee domain.Employee
	if err := r.db.Where("emp_id = ?", empID).First(&employee).Error; err != nil {
		return nil, err
	}

	var pr domain.PullRequest
	err := r.db.
		Where("employee_id = ? AND status IN ?", employee.ID, []string{"open", "draft"}).
		Order("updated_at DESC").
		First(&pr).Error

	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PrRepoImpl) UpdatePRDetails(prID int64, data *dto.GitHubPRResponse) error {
	return r.db.Model(&domain.PullRequest{}).
		Where("id = ?", prID).
		Updates(map[string]interface{}{
			"title":         data.Title,
			"description":   data.Body,
			"status":        data.State,
			"files_changed": data.ChangedFiles,
			"lines_added":   data.Additions,
			"lines_removed": data.Deletions,
			"commit_count":  data.Commits,
			"branch_name":   data.Head.Ref,
			"updated_at":    time.Now(),
		}).Error
}

// Fetch all PRSnapshots for an employee for today
func (r *PrRepoImpl) GetTodaySnapshotsByEmpID(empID string) ([]domain.PRSnapshot, error) {
	var employee domain.Employee
	if err := r.db.Where("emp_id = ?", empID).First(&employee).Error; err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	var snaps []domain.PRSnapshot
	err := r.db.Where("employee_id = ? AND date = ?", employee.ID, today).Find(&snaps).Error
	if err != nil {
		return nil, err
	}
	return snaps, nil
}

// Update SavePRSnapshot to avoid duplicate snapshot for same employee, PR, and date
func (r *PrRepoImpl) SavePRSnapshot(snapshot *domain.PRSnapshot) error {
	today := snapshot.Date.Truncate(24 * time.Hour)
	var existing domain.PRSnapshot
	err := r.db.Where("employee_id = ? AND pr_id = ? AND date = ?", snapshot.EmployeeID, snapshot.PRID, today).First(&existing).Error
	if err == nil {
		// Already exists, update it instead
		snapshot.ID = existing.ID
		return r.db.Model(&existing).Updates(snapshot).Error
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	// Not found, create new
	return r.db.Create(snapshot).Error
}

func (r *PrRepoImpl) GetLatestSnapshot(prID uint) (*domain.PRSnapshot, error) {
	var snap domain.PRSnapshot
	err := r.db.
		Where("pr_id = ?", prID).
		Order("date DESC").
		First(&snap).Error
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

func (r *PrRepoImpl) BuildReportFromSnapshot(pr *domain.PullRequest, snap *domain.PRSnapshot) string {
	return fmt.Sprintf(
		`Daily Update:
Name: %s  
Employee ID: %d
Pull Request: %s  
Branch: %s  
Title: %s
Description: %s

Summary:
Files Changed: %d  
Lines Added: %d  
Lines Removed: %d  
Commits: %d
`,
		snap.Name,
		pr.EmployeeID,
		pr.PRLink,
		pr.BranchName,
		strings.TrimSpace(pr.Title),
		pr.Description,
		snap.FilesChanged,
		snap.LinesAdded,
		snap.LinesRemoved,
		snap.CommitCount,
	)
}

func (r *PrRepoImpl) SavePRReport(report *domain.PRReport) error {
	return r.db.Create(report).Error
}

func (r *PrRepoImpl) FetchReportsByStaffID(empIDs []string) ([]domain.PRReport, error) {
	var reports []domain.PRReport
	today := time.Now().Truncate(24 * time.Hour)
	nextDay := today.Add(24 * time.Hour)
	err := r.db.Table("pr_reports").
		Where("emp_id IN ? AND created_at >= ? AND created_at < ?", empIDs, today, nextDay).
		Order("emp_id, created_at").
		Find(&reports).Error
	return reports, err
}

func (s *PrRepoImpl) BuildMailBody(reports []domain.PRReport) string {
	var sb strings.Builder
	sb.WriteString("Hello,\n\nHere is the PR Report Summary for today:\n\n")

	for _, r := range reports {
		//sb.WriteString(fmt.Sprintf(" Employee ID: %s\n", r.EmpID))
		//sb.WriteString(fmt.Sprintf(" PR Link: %s\n", r.PRLink))
		//sb.WriteString(fmt.Sprintf(" Status: %s\n", r.Status))
		sb.WriteString(fmt.Sprintf(" Summary: %s\n", r.ReportText))
		sb.WriteString("\n---\n\n")
	}

	sb.WriteString("Regards,\nAdmin")
	return sb.String()
}

func (r *PrRepoImpl) GetAllOpenOrDraftPRsByEmpID(empID string) ([]domain.PullRequest, error) {
	var employee domain.Employee
	if err := r.db.Where("emp_id = ?", empID).First(&employee).Error; err != nil {
		return nil, err
	}

	var prs []domain.PullRequest
	err := r.db.
		Where("employee_id = ? AND status IN ?", employee.ID, []string{"open", "draft"}).
		Order("updated_at DESC").
		Find(&prs).Error

	if err != nil {
		return nil, err
	}

	return prs, nil
}

// Fetch a PullRequest by its ID
func (r *PrRepoImpl) GetPRByID(prID uint) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	err := r.db.Table("pull_requests").Where("id = ?", prID).First(&pr).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PrRepoImpl) ReportExistsForEmpAndDate(empID string, date time.Time) (bool, error) {
	var count int64
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	err := r.db.Table("pr_reports").Where("emp_id = ? AND created_at >= ? AND created_at < ?", empID, start, end).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PrRepoImpl) ReportExistsForEmpAndDateAndPR(empID string, date time.Time, prLink string) (bool, error) {
	var count int64
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	err := r.db.Table("pr_reports").Where("emp_id = ? AND pr_link = ? AND created_at >= ? AND created_at < ?", empID, prLink, start, end).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Mark reports as mailed after sending
func (r *PrRepoImpl) MarkReportsAsMailed(reportIDs []uint) error {
	return r.db.Table("pr_reports").Where("id IN ?", reportIDs).Update("is_mail_sent", true).Error
}
