package storage

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	mountPointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"
	"go.uber.org/zap"
)

type batchCreateTextRequest struct {
	Content           string `json:"content" binding:"required"`
	CloudToken        int64  `json:"cloudToken" binding:"required"`
	EnableAutoRefresh bool   `json:"enableAutoRefresh"`
	RefreshInterval   int    `json:"refreshInterval"`
	ShareAccessCode   string `json:"shareAccessCode"`
}

type batchCreateTextResponse struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

var (
	reFolderID  = regexp.MustCompile(`^\d+$`)
	reShareLink = regexp.MustCompile(`cloud\.189\.cn\/t\/([a-zA-Z0-9]+)`)
)

func (h *handler) BatchCreateFromText() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(batchCreateTextRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		tokenInfo, err := h.cloudTokenService.Query(ctx.GetContext(), req.CloudToken)
		if err != nil || tokenInfo == nil {
			ctx.Fail(busCodeStorageCloudTokenNotExist)
			return
		}

		if req.RefreshInterval <= 0 {
			req.RefreshInterval = 30
		}

		lines := strings.Split(req.Content, "\n")
		successCount := 0
		failCount := 0

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.Fields(line)
			resourceStr := parts[0]
			currentAccessCode := req.ShareAccessCode
			if len(parts) > 1 {
				currentAccessCode = parts[1]
			}

			// 1. 先构建 Model 对象
			mp := &models.MountPoint{
				TokenId:           req.CloudToken,
				EnableAutoRefresh: req.EnableAutoRefresh,
				RefreshInterval:   req.RefreshInterval,
			}

			if matches := reShareLink.FindStringSubmatch(resourceStr); len(matches) > 1 {
				shareCode := matches[1]
				mp.Name = "分享导入_" + shareCode
				mp.OsType = protocolSubscribeShare
				mp.TokenName = shareCode
				mp.Password = currentAccessCode
			} else if reFolderID.MatchString(resourceStr) {
				fileId, _ := strconv.ParseInt(resourceStr, 10, 64)
				mp.Name = "文件夹导入_" + resourceStr
				mp.OsType = protocolSubscribe
				mp.FileId = fileId
			} else {
				failCount++
				continue
			}

			// 2. 构造 Request 对象
			createReq := &mountPointSvi.CreateRequest{
				MountPoint: mp,
			}

			// 3. 调用 Service 层创建
			if _, err := h.mountPointService.Create(ctx.GetContext(), createReq); err != nil {
				ctx.GetContext().Error("批量创建失败", zap.String("source", line), zap.Error(err))
				failCount++
			} else {
				successCount++
			}
		}

		ctx.Success(&batchCreateTextResponse{
			Total:   len(lines),
			Success: successCount,
			Failed:  failCount,
		})
	}
}
