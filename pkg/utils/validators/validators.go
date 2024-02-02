package validators

import (
	"errors"
	"regexp"
)

func NonBlankValidate(input string) error {
	if len(input) == 0 {
		return errors.New("cannot be blank")
	}
	return nil
}

func EmailValidate(email string) error {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if matched, _ := regexp.MatchString(emailRegex, email); !matched {
		return errors.New("invalid email format")
	}
	return nil
}

func PasswordValidate(password string) error {
	if len(password) < 5 {
		return errors.New("password must have at least 5 characters")
	}
	return nil
}

func LengthRangeValidate(input string, min, max int) error {
	if len(input) < min || len(input) > max {
		return errors.New("length must be between " + string(min) + " and " + string(max))
	}
	return nil
}

func TextLengthValidate(text string, maxLength int) error {
	if len(text) > maxLength {
		return errors.New("text is too long, maximum length is " + string(maxLength))
	}
	return nil
}
