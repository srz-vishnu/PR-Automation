package e

// Application wide codes
const (
	ErrCodeAuto            = 0
	ErrCodeInternalService = 666
)

// 400 errors
const (
	// ErrInvalidRequest : when post body, query param, or path
	// param is invalid, or any post body validation error is encountered
	ErrInvalidRequest int = 400000 + iota

	// ErrDecodeRequestBody : error when decode the request body
	ErrDecodeRequestBody

	// ErrValidateRequest : error when validating the request
	ErrValidateRequest

	// ErrEmployeeNotValid : error when given employee is not on the table
	ErrEmployeeNotValid

	// ErrSaveUserDetails : error when saving user details
	ErrInvaliPassword

	// ErrTokenNotGenerated : error when getting author by id
	ErrTokenNotGenerated

	// ErrTokenNotSaved : error when not able to save token aginst the user
	ErrTokenNotSaved

	// ErrAdminNotActive : error when not able to save token aginst the user
	ErrAdminNotActive

	// ErrPrCheckingFailed : error when not able to check that pr is exsisting or not
	ErrPrCheckingFailed

	// ErrUpdatePr : error when not able to update pr_link field in the table
	ErrUpdatePr

	// ErrSavePr : error when not able to save pr, name, etc to the table
	ErrSavePr

	// ErrPrParse : error when not able to parse the pr to get basic details
	ErrPrParse

	// ErrPrParse : error when not able to get the pr details from the github
	ErrGitHubAPI

	// ErrUpdatingPRDetails : error when we try to update pr details to the table
	ErrUpdatingPRDetails

	// ErrSavingPRReport : error when we try to save pr report to the table
	ErrSavingPRReport

	// ErrFetchingPRReport : error when we try to fetch report for sending mail.
	ErrFetchingPRReport

	// ErrFetchingPRReport : error when we try to sending  mail.
	ErrSendMail
)

// 404 errors
const (
	// ErrResourceNotFound : when no record corresponding to the requested id is found in the DB
	ErrResourceNotFound int = 404000 + iota
)

// 500 errors
const (
	// ErrInternalServer : the default error, which is unexpected from the developers
	ErrInternalServer int = 500000 + iota

	// ErrExecuteSQL : when execute the sql, meet unexpected error
	ErrExecuteSQL
)
