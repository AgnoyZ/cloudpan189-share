package storage

import (
	"encoding/json"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"

	mountPointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"

	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"

	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/datatypes"

	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"gorm.io/gorm"
)

type (
	addRequest struct {
		LocalPath       string `json:"localPath" binding:"required" example:"/test"`
		OsType          string `json:"osType" binding:"required,oneof=subscribe subscribe_share_folder share_folder person_folder family_folder" example:"subscribe"`
		SubscribeUser   string `json:"subscribeUser" example:"user123"`
		ShareCode       string `json:"shareCode" example:"abc123"`
		ShareAccessCode string `json:"shareAccessCode" example:"1234"`
		CloudToken      int64  `json:"cloudToken" example:"1"`
		FileId          string `json:"fileId" example:"file123"`
		FamilyId        string `json:"familyId" example:"family123"`
	}

	addResponse struct {
		ID   int64  `json:"id" example:"1001"`    // 存储ID
		Path string `json:"path" example:"/test"` // 存储路径
	}
)

// Add 添加存储挂载
// @Summary 添加存储挂载
// @Description 添加新的存储挂载点，支持多种协议类型（订阅、分享、个人云盘、家庭云盘等）
// @Tags 存储管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body addRequest true "存储挂载信息"
// @Success 200 {object} httpcontext.Response{data=addResponse} "存储挂载创建成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "路径分割失败，code=4001"
// @Failure 400 {object} httpcontext.Response "不允许挂载根路径，code=4002"
// @Failure 400 {object} httpcontext.Response "路径不合法，需要 / 开头的路径，code=4003"
// @Failure 400 {object} httpcontext.Response "路径已存在，code=4004"
// @Failure 400 {object} httpcontext.Response "查询路径失败，code=4005"
// @Failure 400 {object} httpcontext.Response "订阅用户不能为空，code=4006"
// @Failure 400 {object} httpcontext.Response "订阅分享参数不完整，code=4008"
// @Failure 400 {object} httpcontext.Response "分享码不能为空，code=4009"
// @Failure 400 {object} httpcontext.Response "个人云盘参数不完整，code=4010"
// @Failure 400 {object} httpcontext.Response "家庭云盘参数不完整，code=4011"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/storage/add [post]
func (h *handler) Add() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(addRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)

			return
		}

		if mountPoint, err := h.mountPointService.QueryByPath(ctx.GetContext(), req.LocalPath); err != nil {
			ctx.Fail(busCodeStorageQueryPathFailed.WithError(err))

			return
		} else if mountPoint != nil {
			ctx.Fail(busCodeStoragePathExists)

			return
		}

		paths, err := utils.SplitPath(req.LocalPath)
		if err != nil {
			ctx.Fail(busCodeStoragePathSplitFailed.WithError(err))

			return
		}

		if len(paths) == 0 {
			ctx.Fail(busCodeStorageRootPathNotAllowed)

			return
		}

		if !utils.CheckIsPath(req.LocalPath) {
			ctx.Fail(busCodeStorageInvalidPath)

			return
		}

		if _, err = h.virtualFileService.QueryByPath(ctx.GetContext(), req.LocalPath); !errors.Is(err, gorm.ErrRecordNotFound) {
			if err != nil {
				ctx.Fail(busCodeStorageQueryPathFailed.WithError(err))
			} else {
				ctx.Fail(busCodeStoragePathExists)
			}

			return
		}

		var (
			addition datatypes.JSONMap
		)

		fileId := req.FileId

		switch req.OsType {
		case protocolSubscribe:
			addition, err = h.executeOsTypeSubscribe(ctx.GetContext(), req)
		case protocolSubscribeShare:
			addition, fileId, err = h.executeOsTypeSubscribeShare(ctx.GetContext(), req)
		case protocolShare:
			addition, fileId, err = h.executeOsTypeShare(ctx.GetContext(), req)
		case protocolPerson:
			err = h.executeOsTypePersonal(ctx.GetContext(), req)
			addition = datatypes.JSONMap{}
		case protocolFamily:
			addition, err = h.executeOsTypeFamily(ctx.GetContext(), req)
		default:
			ctx.Fail(busCodeStorageOsTypeUnsupport)

			return
		}

		if err != nil {
			ctx.Fail(busCodeStorageQueryPathFailed.WithError(err))

			return
		}

		// 查找或创建父级路径
		parentId, err := h.virtualFileService.FindOrCreateAncestors(ctx.GetContext(), req.LocalPath)
		if err != nil {
			ctx.Fail(busCodeStorageQueryPathFailed.WithError(err))

			return
		}

		now := time.Now()
		m := &models.VirtualFile{
			Name:       paths[len(paths)-1],
			IsTop:      true,
			Size:       0,
			Hash:       "",
			CreateDate: now,
			ModifyDate: now,
			Rev:        now.Format(consts.RevFormat),
			OsType:     req.OsType,
			IsDir:      true,
			Addition:   addition,
			CloudId:    fileId,
		}

		// 创建存储挂载
		id, err := h.virtualFileService.CreateTop(ctx.GetContext(), parentId, m)
		if err != nil {
			ctx.Fail(busCodeStorageQueryPathFailed.WithError(err))

			return
		}

		if _, err = h.mountPointService.Create(ctx.GetContext(), &mountPointSvi.CreateRequest{
			FullPath: req.LocalPath,
			FileId:   id,
			OsType:   req.OsType,
			TokenId:  req.CloudToken,
		}); err != nil {
			ctx.Fail(busCodeStorageAddMountPointFailed.WithError(err))

			return
		}

		taskReq := &topic.FileScanFileRequest{
			FileId: id,
			Deep:   true,
		}

		body, _ := json.Marshal(taskReq)
		if err = h.taskEngine.PushMessage(ctx.GetContext().WithValue(consts.CtxKeyFullPath, req.LocalPath), taskReq.Topic(), body); err != nil {
			ctx.Fail(busCodeStorageAddTaskFailed.WithError(err))

			return
		}

		ctx.Success(&addResponse{
			ID:   id,
			Path: req.LocalPath,
		})
	}
}
