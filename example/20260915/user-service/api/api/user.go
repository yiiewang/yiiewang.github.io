package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/api/forms"
	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/api/global"
	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/api/global/response"
	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/proto"
)

// removeTopStruct 把 "PasswordLoginForm.Mobile" 这类校验键裁成 "mobile"
func removeTopStruct(fields map[string]string) map[string]string {
	rsp := map[string]string{}
	for field, err := range fields {
		rsp[field[strings.Index(field, ".")+1:]] = err
	}
	return rsp
}

// HandleGrpcErrorToHttp 把 gRPC status code 映射成 HTTP 状态码
func HandleGrpcErrorToHttp(err error, ctx *gin.Context) {
	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.NotFound:
				ctx.JSON(http.StatusNotFound, gin.H{
					"msg": s.Message(),
				})
			case codes.Internal:
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"msg": "内部错误",
				})
			case codes.InvalidArgument:
				ctx.JSON(http.StatusBadRequest, gin.H{
					"msg": "参数错误",
				})
			case codes.AlreadyExists:
				ctx.JSON(http.StatusBadRequest, gin.H{
					"msg": s.Message(),
				})
			default:
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"msg": "其他错误 ",
				})
			}
		}
	}
}

// GetUserList 用户列表
func GetUserList(ctx *gin.Context) {
	uc := proto.NewUserClient(global.UserSrvConn)

	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "0"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	zap.S().Debug(pageNum)
	zap.S().Debug(pageSize)

	rsp, err := uc.GetUserList(context.Background(), &proto.PageInfo{
		PageNum:  uint32(pageNum),
		PageSize: uint32(pageSize),
	})
	if err != nil {
		zap.S().Errorw("GetUserList visited user server failed", "msg", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	result := make([]response.UserResponse, 0)
	for _, v := range rsp.Data {
		data := response.UserResponse{
			Id:       v.Id,
			NickName: v.NickName,
			Birthday: time.Unix(int64(v.Birthday), 0).Format("2006-01-02"),
			Gender:   v.Gender,
			Mobile:   v.Mobile,
		}
		result = append(result, data)
	}

	// 注意：不能把 rsp 原样 JSON 出去，UserInfoResponse 里有密码哈希字段
	ctx.JSON(http.StatusOK, gin.H{
		"total": rsp.Total,
		"data":  result,
	})
}

// Register 注册用户：HTTP 表单 → CreateUser RPC（服务端做 pbkdf2-sha512 加密）
func Register(ctx *gin.Context) {
	registerForm := forms.RegisterForm{}
	if err := ctx.ShouldBind(&registerForm); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": removeTopStruct(errs.Translate(global.Trans)),
		})
		return
	}

	uc := proto.NewUserClient(global.UserSrvConn)

	userRsp, err := uc.CreateUser(context.Background(), &proto.CreateUserInfo{
		NickName: registerForm.NickName,
		Mobile:   registerForm.Mobile,
		Password: registerForm.Password,
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id": userRsp.Id,
	})
}

// PasswordLogin 密码登录：先按手机号查用户拿到密文，再走 CheckPasswd 校验
func PasswordLogin(ctx *gin.Context) {
	passwordLoginForm := forms.PasswordLoginForm{}
	if err := ctx.ShouldBind(&passwordLoginForm); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": removeTopStruct(errs.Translate(global.Trans)),
		})
		return
	}

	uc := proto.NewUserClient(global.UserSrvConn)

	// 1. 通过 mobile 查询用户（拿加密后的密码）
	userRsp, err := uc.GetUserByMobile(context.Background(), &proto.MobileRequest{
		Mobile: passwordLoginForm.Mobile,
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	// 2. 校验明文密码与密文是否匹配
	checkRsp, err := uc.CheckPasswd(context.Background(), &proto.PasswdCheckInfo{
		Password:          passwordLoginForm.Password,
		EncryptedPassword: userRsp.Password,
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	if !checkRsp.Success {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "密码错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "登录成功",
	})
}
