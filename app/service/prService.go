package service

import (
	"errors"
	"fmt"
	"net/http"
	"pr-mail/app/domain"
	"pr-mail/app/dto"
	github "pr-mail/app/github"
	helper "pr-mail/app/helper"
	"pr-mail/app/repo"
	"pr-mail/pkg/e"
	"pr-mail/pkg/jwt"
	"pr-mail/pkg/smtp"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type PrService interface {
	LoginUser(r *http.Request) (*dto.LoginResponse, error)
	SaveEmployeePR(r *http.Request) error
	GeneratePRDetails(r *http.Request) (*dto.PRDetailsResponse, error)
	GeneratePRReport(r *http.Request) (*dto.PRReportResponse, error)
	SendPRMail(r *http.Request) error
}

type prServiceImpl struct {
	prRepo repo.PrRepo
}

func NewPrService(prRepo repo.PrRepo) PrService {
	return &prServiceImpl{
		prRepo: prRepo,
	}
}

func (s *prServiceImpl) LoginUser(r *http.Request) (*dto.LoginResponse, error) {
	args := &dto.LoginRequest{}

	// parsing the req.body
	err := args.Parse(r)
	if err != nil {
		return nil, e.NewError(e.ErrDecodeRequestBody, "error while parsing", err)
	}

	//validation
	err = args.Validate()
	if err != nil {
		return nil, e.NewError(e.ErrValidateRequest, "error while validating", err)
	}
	log.Info().Msg("Successfully completed parsing and validation of request body")

	// Fetching user from database
	details, err := s.prRepo.GetUserByUsername(args.Username)
	if err != nil {
		return nil, e.NewError(e.ErrResourceNotFound, "user not found", err)
	}

	// Check if details is nil
	if details == nil {
		return nil, e.NewError(e.ErrResourceNotFound, "user not found", err)
	}
	log.Info().Msgf("the admin is %s", details.Username)

	if !details.Status {
		return nil, e.NewError(e.ErrAdminNotActive, "admin is not active", nil)
	}

	// Validate password
	if details.Password != args.Password {
		err := fmt.Errorf("invalid password for user %s", details.Username)
		return nil, e.NewError(e.ErrInvaliPassword, "invalid password", err)
	}

	// Generating JWT Token
	token, err := jwt.GenerateToken(int64(details.ID), details.Username)
	if err != nil {
		return nil, e.NewError(e.ErrTokenNotGenerated, "failed to generate token", err)
	}

	fmt.Printf("the token is %s : \n ", token)

	// Save token to DB
	err = s.prRepo.UpdateUserToken(details.ID, token)
	if err != nil {
		return nil, e.NewError(e.ErrTokenNotSaved, "failed to save token", err)
	}

	return &dto.LoginResponse{
		Token: token,
	}, nil
}

func (s *prServiceImpl) SaveEmployeePR(r *http.Request) error {
	args := &dto.SaveEmployeePRRequest{}

	// Parse request body
	err := args.Parse(r)
	if err != nil {
		return e.NewError(e.ErrDecodeRequestBody, "error while parsing", err)
	}

	//  Validate input
	err = args.Validate()
	if err != nil {
		return e.NewError(e.ErrValidateRequest, "error while validating", err)
	}
	log.Info().Msg("Successfully parsed and validated SaveEmployeePR request")

	//  Check if employee exists
	employee, err := s.prRepo.GetEmployeeByEmpID(args.StaffID)
	if err != nil {
		return e.NewError(e.ErrResourceNotFound, "employee not found", err)
	}

	// Check if PR already exists for this employee
	existingPR, err := s.prRepo.GetPRByEmpIDAndLink(employee.ID, args.PRLink)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return e.NewError(e.ErrPrCheckingFailed, "failed to check existing PR", err)
	}

	now := time.Now()

	if existingPR != nil {
		//Update `updated_at` and pr_link (if needed)
		existingPR.PRLink = args.PRLink
		existingPR.UpdatedAt = &now
		err := s.prRepo.UpdatePullRequest(existingPR)
		if err != nil {
			return e.NewError(e.ErrUpdatePr, "failed to update existing PR", err)
		}
		log.Info().Msg("PR already existed, updated timestamp")
		return nil
	}

	// new PR entry
	newPR := &domain.PullRequest{
		EmployeeID: employee.ID,
		StaffID:    args.StaffID,
		PRLink:     args.PRLink,
		Status:     args.Status,
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}
	err = s.prRepo.SavePullRequest(newPR)
	if err != nil {
		return e.NewError(e.ErrSavePr, "failed to save new PR", err)
	}

	log.Info().Msg("Successfully saved new pull request")
	return nil
}

func (s *prServiceImpl) GeneratePRDetails(r *http.Request) (*dto.PRDetailsResponse, error) {
	args := &dto.PRDetailsEmployeeID{}

	// Parse request body
	err := args.Parse(r)
	if err != nil {
		return nil, e.NewError(e.ErrDecodeRequestBody, "error while parsing", err)
	}
	log.Info().Msg("Successfully parsed request")

	err = args.Validate()
	if err != nil {
		return nil, e.NewError(e.ErrValidateRequest, "error while validating", err)
	}
	log.Info().Msg("Successfully completed parsing and validation of request body")

	//  validating employee is there on db
	pr, err := s.prRepo.ValidPRByEmpID(args.StaffID)
	if err != nil {
		return nil, e.NewError(e.ErrResourceNotFound, "No recent PR found for this employee", err)
	}
	log.Info().Msg("Successfully got pr")

	//  Extract repo owner/name and PR number from pr.PRLink
	owner, repo, prNumber, err := helper.ParsePRLink(pr.PRLink)
	if err != nil {
		return nil, e.NewError(e.ErrPrParse, "Invalid PR link format", err)
	}
	log.Info().Msg("Successfully extracted pr basic details")
	log.Info().Msgf("owner %s:", owner)
	log.Info().Msgf("pr num %d:", prNumber)
	log.Info().Msgf("repo %s:", repo)

	//  GitHub API to fetch PR details
	prData, err := github.FetchPRDetails(owner, repo, prNumber)
	// prData, err := github.FetchPRDetails(pr.PRLink)
	if err != nil {
		return nil, e.NewError(e.ErrGitHubAPI, "Failed to fetch PR data from GitHub", err)
	}
	log.Info().Msg("Successfully get github api details")
	log.Info().Msgf("name is%s:", prData.Body)

	// Save today's PR data to a table so we can see the daily chnage here
	snapshot := &domain.PRSnapshot{
		EmployeeID:   pr.EmployeeID,
		Name:         owner,
		PRID:         pr.ID,
		Date:         time.Now().Truncate(24 * time.Hour),
		Description:  pr.Description,
		LinesAdded:   prData.Additions,
		LinesRemoved: prData.Deletions,
		FilesChanged: prData.ChangedFiles,
		CommitCount:  prData.Commits,
	}

	err = s.prRepo.SavePRSnapshot(snapshot)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save daily PR details")
	}

	err = s.prRepo.UpdatePRDetails(int64(pr.ID), prData)
	if err != nil {
		return nil, e.NewError(e.ErrUpdatingPRDetails, "failed to update PR details in DB", err)
	}
	log.Info().Msg("PR Details Saved")

	//  responseeee
	return &dto.PRDetailsResponse{
		Owner:        owner,
		Title:        prData.Title,
		Description:  prData.Body,
		Status:       prData.State,
		Files:        prData.ChangedFiles,
		LinesAdded:   prData.Additions,
		LinesRemoved: prData.Deletions,
		CommitCount:  prData.Commits,
		Branch:       prData.Head.Ref,
	}, nil
}

func (s *prServiceImpl) GeneratePRReport(r *http.Request) (*dto.PRReportResponse, error) {
	args := &dto.PRReportRequest{}

	// Parse request body
	err := args.Parse(r)
	if err != nil {
		return nil, e.NewError(e.ErrDecodeRequestBody, "error while parsing", err)
	}
	log.Info().Msg("Successfully parsed request")

	err = args.Validate()
	if err != nil {
		return nil, e.NewError(e.ErrValidateRequest, "error while validating", err)
	}
	log.Info().Msg("Successfully completed parsing and validation of request body")

	//  checkingg employee is there on db and getting the pr details also
	pr, err := s.prRepo.ValidPRByEmpID(args.StaffID)
	if err != nil {
		return nil, e.NewError(e.ErrResourceNotFound, "No recent PR found for this employee", err)
	}
	log.Info().Msg("Successfully got pr details from table")

	snap, err := s.prRepo.GetLatestSnapshot(pr.ID)
	if err != nil {
		return nil, e.NewError(e.ErrResourceNotFound, "No snapshot found for PR", err)
	}
	log.Info().Msg("Successfully got daily PR detials")

	reportText := s.prRepo.BuildReportFromSnapshot(pr, snap)

	// Saving the report in reports table
	report := &domain.PRReport{
		EmpID:      args.StaffID,
		PRLink:     pr.PRLink,
		Status:     pr.Status,
		ReportText: reportText,
		IsMailSent: false,
	}
	if err := s.prRepo.SavePRReport(report); err != nil {
		return nil, e.NewError(e.ErrSavingPRReport, "Failed to store PR report", err)
	}

	fmt.Println(reportText)

	return &dto.PRReportResponse{
		EmpID:      args.StaffID,
		PRLink:     pr.PRLink,
		ReportText: reportText,
	}, nil
}

func (s *prServiceImpl) SendPRMail(r *http.Request) error {
	args := &dto.SendMailRequest{}

	// Parse request body
	err := args.Parse(r)
	if err != nil {
		return e.NewError(e.ErrDecodeRequestBody, "error while parsing", err)
	}
	log.Info().Msg("Successfully parsed request")

	if len(args.StaffID) == 0 {
		return e.NewError(e.ErrValidateRequest, "employee list is empty", nil)
	}
	log.Info().Msg("Successfully completed parsing and validation of request body")

	// Fetch reports
	reports, err := s.prRepo.FetchReportsByStaffID(args.StaffID)
	if err != nil {
		return e.NewError(e.ErrFetchingPRReport, "Failed to fetch PR reports", err)
	}
	log.Info().Msgf("Successfully fetched reports: %d", len(reports))

	// Build mail body
	body := s.prRepo.BuildMailBody(reports)

	// Send email
	err = smtp.SendEmail(body)
	if err != nil {
		return e.NewError(e.ErrSendMail, "Failed to send mail", err)
	}
	//status true

	return nil
}
