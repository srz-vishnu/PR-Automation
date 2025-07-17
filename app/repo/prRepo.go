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

func (r *PrRepoImpl) SavePRSnapshot(snapshot *domain.PRSnapshot) error {
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

	subQuery := r.db.Table("pr_reports").
		Select("MAX(updated_at)").
		Where("emp_id = pr.emp_id AND is_mail_sent = false")

	err := r.db.Table("pr_reports AS pr").
		Where("pr.emp_id IN ? AND pr.updated_at = (?)", empIDs, subQuery).
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
