package errors

var (
	ErrOutOfDailyToken = &InternalError{
		Code:    1,
		Message: "out of daily token",
	}
	ErrUserNotFound = &InternalError{
		Code:    2,
		Message: "user not found",
	}
	ErrFailedToAnalyze = &InternalError{
		Code:    3,
		Message: "failed to analyze",
	}
	ErrFaiedToGetUser = &InternalError{
		Code:    4,
		Message: "failed to get user",
	}
	ErrFaiedToGetImage = &InternalError{
		Code:    5,
		Message: "failed to get image",
	}
	ErrNotFoodType = &InternalError{
		Code:    6,
		Message: "not food type",
	}
	ErrInternal = &InternalError{
		Code:    7,
		Message: "internal error",
	}
	ErrRegister = &InternalError{
		Code:    8,
		Message: "register error",
	}
	ErrNoHistory = &InternalError{
		Code:    9,
		Message: "no history",
	}
	ErrEndDateBeforeStartDate = &InternalError{
		Code:    10,
		Message: "end date before start date",
	}
)

type InternalError struct {
	Code    int
	Message string
}

func (e *InternalError) Error() string {
	return e.Message
}
