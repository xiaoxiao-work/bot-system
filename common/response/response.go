package response

import "github.com/gin-gonic/gin"

// 统一响应结构
type Response struct {
	ErrCode int         `json:"errCode"`
	ErrMsg  string      `json:"errMsg"`
	ErrDlt  string      `json:"errDlt,omitempty"`
	Data    interface{} `json:"data"`
}

// 错误码定义
const (
	Success               = 0
	ErrBadRequest         = 400
	ErrNotFound           = 404
	ErrInternalServer     = 500
	ErrBotNotFound        = 1001
	ErrBotAlreadyExist    = 1002
	ErrRegisterFailed     = 1003
	ErrGetTokenFailed     = 1004
	ErrInviteFailed       = 1005
	ErrKickFailed         = 1006
	ErrSendMessageFailed  = 1007
)

// 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		ErrCode: Success,
		ErrMsg:  "success",
		Data:    data,
	})
}

// 成功响应（带自定义消息）
func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(200, Response{
		ErrCode: Success,
		ErrMsg:  msg,
		Data:    data,
	})
}

// 错误响应
func ErrorResponse(c *gin.Context, errCode int, errMsg string) {
	c.JSON(200, Response{
		ErrCode: errCode,
		ErrMsg:  errMsg,
		Data:    nil,
	})
}

// 错误响应（带详细信息）
func ErrorWithDetail(c *gin.Context, errCode int, errMsg string, errDlt string) {
	c.JSON(200, Response{
		ErrCode: errCode,
		ErrMsg:  errMsg,
		ErrDlt:  errDlt,
		Data:    nil,
	})
}

// 错误响应（从 error 对象）
func ErrorFromErr(c *gin.Context, errCode int, errMsg string, err error) {
	var errDlt string
	if err != nil {
		errDlt = err.Error()
	}
	c.JSON(200, Response{
		ErrCode: errCode,
		ErrMsg:  errMsg,
		ErrDlt:  errDlt,
		Data:    nil,
	})
}
