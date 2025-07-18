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

	// Parse request
	err := args.Parse(r)
	if err != nil {
		return e.NewError(e.ErrDecodeRequestBody, "error while parsing", err)
	}

	// Validate
	err = args.Validate()
	if err != nil {
		return e.NewError(e.ErrValidateRequest, "error while validating", err)
	}
	log.Info().Msg("Successfully parsed and validated SaveEmployeePR request")

	now := time.Now()

	// Process each PR
	for _, pr := range args.PRs {
		employee, err := s.prRepo.GetEmployeeByEmpID(pr.StaffID)
		if err != nil {
			log.Error().Err(err).Msgf("Employee not found for staff ID %s", pr.StaffID)
			return e.NewError(e.ErrResourceNotFound, "employee not found", err)
			//continue // Skip and continue with other PRs
		}

		existingPR, err := s.prRepo.GetPRByEmpIDAndLink(employee.ID, pr.PRLink)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return e.NewError(e.ErrPrCheckingFailed, "failed to check existing PR", err)
		}

		if existingPR != nil {
			existingPR.PRLink = pr.PRLink
			existingPR.UpdatedAt = &now
			err := s.prRepo.UpdatePullRequest(existingPR)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to update existing PR for %s", pr.PRLink)
				return e.NewError(e.ErrUpdatePr, "failed to update existing PR", err)
			} else {
				log.Info().Msgf("Updated existing PR for %s", pr.PRLink)
			}
			continue
		}

		// Save new PR
		newPR := &domain.PullRequest{
			EmployeeID: employee.ID,
			StaffID:    pr.StaffID,
			PRLink:     pr.PRLink,
			Status:     pr.Status,
			CreatedAt:  &now,
			UpdatedAt:  &now,
		}

		err = s.prRepo.SavePullRequest(newPR)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to save new PR for %s", pr.PRLink)
			return e.NewError(e.ErrSavePr, "failed to save new PR", err)
			//continue
		}

		log.Info().Msgf("Successfully saved new PR for %s", pr.PRLink)
	}

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

	// Fetch all PRs for this employee with status open or draft
	prs, err := s.prRepo.GetAllOpenOrDraftPRsByEmpID(args.StaffID)
	if err != nil || len(prs) == 0 {
		return nil, e.NewError(e.ErrResourceNotFound, "No open/draft PRs found for this employee", err)
	}
	log.Info().Msgf("Found %d open/draft PRs", len(prs))

	var prDetailsList []dto.SinglePRDetails

	for _, pr := range prs {
		// Extract repo owner/name and PR number from pr.PRLink
		owner, repo, prNumber, err := helper.ParsePRLink(pr.PRLink)
		if err != nil {
			log.Error().Err(err).Msgf("Invalid PR link format: %s", pr.PRLink)
			continue // skip this PR
		}
		log.Info().Msgf("Processing PR: owner=%s, repo=%s, prNumber=%d", owner, repo, prNumber)

		// GitHub API to fetch PR details
		prData, err := github.FetchPRDetails(owner, repo, prNumber)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to fetch PR data from GitHub for PR: %s", pr.PRLink)
			continue // skip this PR
		}

		// Save today's PR data to a table so we can see the daily change here
		snapshot := &domain.PRSnapshot{
			EmployeeID:   pr.EmployeeID,
			Name:         owner,
			PRID:         pr.ID,
			Date:         time.Now().Truncate(24 * time.Hour),
			Description:  prData.Body, // Use PR description from GitHub
			LinesAdded:   prData.Additions,
			LinesRemoved: prData.Deletions,
			FilesChanged: prData.ChangedFiles,
			CommitCount:  prData.Commits,
		}
		err = s.prRepo.SavePRSnapshot(snapshot)
		if err != nil {
			log.Error().Err(err).Msg("Failed to save daily PR details")
		}

		// Update PR with pr_number, repo_name, source_branch, and description
		pr.PRNumber = prNumber
		pr.RepoName = repo
		pr.BranchName = prData.Head.Ref
		pr.Description = prData.Body
		err = s.prRepo.UpdatePullRequest(&pr)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update PR details in DB")
		}

		prDetailsList = append(prDetailsList, dto.SinglePRDetails{
			Owner:        owner,
			Title:        prData.Title,
			Description:  prData.Body,
			Status:       prData.State,
			Files:        prData.ChangedFiles,
			LinesAdded:   prData.Additions,
			LinesRemoved: prData.Deletions,
			CommitCount:  prData.Commits,
			Branch:       prData.Head.Ref,
			PRLink:       pr.PRLink,
		})
	}

	if len(prDetailsList) == 0 {
		return nil, e.NewError(e.ErrResourceNotFound, "No valid PR details could be fetched for this employee", nil)
	}

	return &dto.PRDetailsResponse{
		PRs: prDetailsList,
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

	// Fetch all today's PRSnapshots for this employee
	snaps, err := s.prRepo.GetTodaySnapshotsByEmpID(args.StaffID)
	if err != nil || len(snaps) == 0 {
		return nil, e.NewError(e.ErrResourceNotFound, "No PR snapshots found for this employee today", err)
	}
	log.Info().Msgf("Found %d PR snapshots for today", len(snaps))

	// Fetch all PRs for these snapshots and save a report for each
	var reportText string
	var savedReports []domain.PRReport
	for _, snap := range snaps {
		pr, err := s.prRepo.GetPRByID(snap.PRID)
		if err != nil {
			log.Error().Err(err).Msgf("No PR found for snapshot PRID: %d", snap.PRID)
			continue
		}
		text := s.prRepo.BuildReportFromSnapshot(pr, &snap)
		reportText += text + "\n---\n"

		// Check if a report for this employee, PR, and today already exists
		today := time.Now().Truncate(24 * time.Hour)
		exists, err := s.prRepo.ReportExistsForEmpAndDateAndPR(args.StaffID, today, pr.PRLink)
		if err != nil {
			return nil, e.NewError(e.ErrSavingPRReport, "Failed to check for existing report", err)
		}
		if exists {
			log.Info().Msgf("Report for %s and PR %s already exists for today, skipping save", args.StaffID, pr.PRLink)
			continue
		}

		report := &domain.PRReport{
			EmpID:      args.StaffID,
			PRLink:     pr.PRLink,
			Status:     pr.Status,
			ReportText: text,
			IsMailSent: false,
		}
		if err := s.prRepo.SavePRReport(report); err != nil {
			return nil, e.NewError(e.ErrSavingPRReport, "Failed to store PR report", err)
		}
		savedReports = append(savedReports, *report)
	}

	if reportText == "" {
		return nil, e.NewError(e.ErrResourceNotFound, "No valid PR reports could be built for this employee today", nil)
	}

	log.Info().Msgf("Generated PR report for %s: \n%s", args.StaffID, reportText)

	return &dto.PRReportResponse{
		EmpID:      args.StaffID,
		PRLink:     "",
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

	if len(reports) == 0 {
		return e.NewError(e.ErrFetchingPRReport, "No unsent PR reports found for these employees", nil)
	}

	// Build mail body
	body := s.prRepo.BuildMailBody(reports)

	// Send email
	err = smtp.SendEmail(body)
	if err != nil {
		return e.NewError(e.ErrSendMail, "Failed to send mail", err)
	}

	// Mark reports as mailed
	reportIDs := make([]uint, 0, len(reports))
	for _, r := range reports {
		reportIDs = append(reportIDs, r.ID)
	}
	err = s.prRepo.MarkReportsAsMailed(reportIDs)
	if err != nil {
		return e.NewError(e.ErrSendMail, "Failed to update mail sent status", err)
	}
	//status true

	return nil
}
