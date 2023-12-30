package app

import (
	"fmt"
	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"message-nest/pkg/e"
	"message-nest/pkg/util"
	"net/http"
	"strings"
)

// BindAndValid binds and validates data
func BindAndValid(c *gin.Context, form interface{}) (int, int) {
	err := c.Bind(form)
	if err != nil {
		return http.StatusBadRequest, e.INVALID_PARAMS
	}

	valid := validation.Validation{}
	check, err := valid.Valid(form)
	if err != nil {
		return http.StatusInternalServerError, e.ERROR
	}
	if !check {
		MarkErrors(valid.Errors)
		return http.StatusBadRequest, e.INVALID_PARAMS
	}

	return http.StatusOK, e.SUCCESS
}

func CommonPlaygroundValid(obj interface{}) (int, string) {
	if err := util.CustomerValidate.Struct(obj); err != nil {
		errs := err.(validator.ValidationErrors)
		errMsg := BuildValidationErrors(errs)
		return http.StatusBadRequest, errMsg
	}
	return http.StatusOK, ""
}

func BindJsonAndPlayValid(c *gin.Context, req interface{}) (int, string) {
	err := c.ShouldBindJSON(req)
	if err != nil {
		return http.StatusBadRequest, err.Error()
	} else {
		return CommonPlaygroundValid(req)
	}
}

func BuildValidationErrors(errors []validator.FieldError) string {
	var errorMsgBuilder strings.Builder
	for i, err := range errors {
		if i > 0 {
			errorMsgBuilder.WriteString("; ")
		}
		message := err.Translate(util.Trans)
		errorMsgBuilder.WriteString(fmt.Sprintf("%s", message))
	}
	return errorMsgBuilder.String()
}
