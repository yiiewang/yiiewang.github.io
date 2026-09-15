package handler

import (
	"context"
	"crypto/sha512"
	"fmt"
	"strings"
	"time"

	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/proto"
	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/srv/global"
	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/srv/model"
)

type UserServer struct {
	proto.UnimplementedUserServer
}

func Model2Response(user model.User) *proto.UserInfoResponse {
	// 在 grpc message 中不能存储 nil
	// 确定哪些字段有默认值
	userInfoRsp := proto.UserInfoResponse{
		Id:       user.ID,
		Password: user.Password,
		Mobile:   user.Mobile,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     int32(user.Role),
	}
	if user.Birthday != nil {
		userInfoRsp.Birthday = uint64(user.Birthday.Unix())
	}
	return &userInfoRsp
}

// GetUserList 获取用户列表
func (s *UserServer) GetUserList(ctx context.Context, req *proto.PageInfo) (*proto.UserListResponse, error) {

	var users []model.User
	result := global.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.UserListResponse{}
	rsp.Total = int32(result.RowsAffected)

	// 分页
	global.DB.Scopes(Paginate(int(req.PageNum), int(req.PageSize))).Find(&users)

	for _, user := range users {
		userInfoRsp := Model2Response(user)
		rsp.Data = append(rsp.Data, userInfoRsp)
	}
	return rsp, nil
}

// GetUserByMobile 通过 mobile 查询用户
func (s *UserServer) GetUserByMobile(ctx context.Context, req *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	d := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if d.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "没有找到用户")
	}
	if d.Error != nil {
		return nil, d.Error
	}

	userInfoRsp := Model2Response(user)
	return userInfoRsp, nil
}

// GetUserByID 通过 id 查询用户
func (s *UserServer) GetUserByID(ctx context.Context, req *proto.IDRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	d := global.DB.First(&user, req.Id)
	if d.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "没有找到用户")
	}
	if d.Error != nil {
		return nil, d.Error
	}

	userInfoRsp := Model2Response(user)
	return userInfoRsp, nil
}

// CreateUser 新建用户
func (s *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	var user model.User
	d := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if d.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "用户已存在")
	}
	user.Mobile = req.Mobile
	user.NickName = req.NickName

	// 密码加密
	options := &password.Options{SaltLen: 10, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	salt, encodedPwd := password.Encode(req.Password, options)
	user.Password = fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)

	d2 := global.DB.Create(&user)
	if d2.Error != nil {
		return nil, status.Error(codes.Internal, d2.Error.Error())
	}

	userInfoRsp := Model2Response(user)
	return userInfoRsp, nil
}

// UpdateUser 更新用户
func (s *UserServer) UpdateUser(ctx context.Context, req *proto.UpdateUserInfo) (*emptypb.Empty, error) {
	var user model.User
	d := global.DB.First(&user, req.Id)
	if d.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}

	birthday := time.Unix(int64(req.Birthday), 0)
	user.NickName = req.NickName
	user.Birthday = &birthday
	user.Gender = req.Gender
	d = global.DB.Save(&user)
	if d.Error != nil {
		return nil, status.Errorf(codes.Internal, d.Error.Error())
	}
	return &emptypb.Empty{}, nil
}

// CheckPasswd 校验明文密码与已加密密码是否匹配
func (s *UserServer) CheckPasswd(ctx context.Context, req *proto.PasswdCheckInfo) (*proto.CheckResponse, error) {
	options := &password.Options{SaltLen: 10, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	splits := strings.Split(req.EncryptedPassword, "$")
	b := password.Verify(req.Password, splits[2], splits[3], options)
	return &proto.CheckResponse{Success: b}, nil
}

func Paginate(pageNum, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pageNum == 0 {
			pageNum = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (pageNum - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
