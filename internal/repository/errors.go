package repository

import "errors"

var ErrorConflict = errors.New("conflict")
var ErrorVersionConflict = errors.New("Error version conflict")
var ErrorDoubleOperation = errors.New("double operation")
var ErrorUserNotFound = errors.New("user not found")
var ErrorNotContent = errors.New("data not found")
var ErrorEmptyUpload = errors.New("data not found")
