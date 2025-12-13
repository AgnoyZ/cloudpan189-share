package storage

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"go.uber.org/zap"
)

// batchCreateTextRequest 接收前端表单参数
// 参考了 func_add.go 中的参数设计，允许用户设置通用的挂载属性
type batchCreateTextRequest struct {
	Content string `json:"content" binding:"required"` // 批量文本内容

	// --- 以下为通用设置，应用到批量创建的每一个挂载点 ---
	CloudToken        int64  `json:"cloudToken" binding:"required"` // 所属账号ID
	EnableAutoRefresh bool   `json:"enableAutoRefresh"`             // 是否开启自动刷新
	RefreshInterval   int    `json:"refreshInterval"`               // 刷新间隔(分钟)
	ShareAccessCode   string `json:"shareAccessCode"`               // 默认访问码(仅对分享链接有效)
}

// 响应结构
type batchCreateTextResponse struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// 预编译正则
var (
	// 匹配纯数字 (Folder ID)
	reFolderID = regexp.MustCompile(`^\d+$`)
	// 匹配天翼云分享链接，提取 /t/ 后面的 Code
	reShareLink = regexp.MustCompile(`cloud\.189\.cn\/t\/([a-zA-Z0-9]+)`)
)

// BatchCreateFromText 批量导入挂载点
// @Summary 批量导入挂载点
// @Description 支持输入多行文本（文件夹ID或分享链接），并应用统一的挂载设置
// @Tags 存储管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body batchCreateTextRequest true "导入设置"
// @Success 200 {object} httpcontext.Response{data=batchCreateTextResponse} "操作成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败"
// @Failure 400 {object} httpcontext.Response "CloudToken不存在"
// @Router /api/storage/batch_create_text [post]
func (h *handler) BatchCreateFromText() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(batchCreateTextRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		// 1. 校验 CloudToken 是否存在 (参考 func_add.go 的逻辑)
		tokenInfo, err := h.cloudTokenService.Query(ctx.GetContext(), req.CloudToken)
		if err != nil {
			ctx.Fail(busCodeStorageCloudTokenNotExist) // 使用 handler.go 里定义的错误码
			return
		}
		if tokenInfo == nil {
			ctx.Fail(busCodeStorageCloudTokenNotExist)
			return
		}

		// 2. 处理默认值
		if req.RefreshInterval <= 0 {
			req.RefreshInterval = 30 // 默认30分钟
		}

		lines := strings.Split(req.Content, "\n")
		successCount := 0
		failCount := 0

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 支持行内访问码格式： "链接 访问码" 或 "ID 访问码"
			// split by space
			parts := strings.Fields(line)
			resourceStr := parts[0]
			currentAccessCode := req.ShareAccessCode // 默认使用全局设置的访问码

			// 如果该行指定了独立的访问码，覆盖全局设置
			if len(parts) > 1 {
				currentAccessCode = parts[1]
			}

			var mp *models.MountPoint

			// 情况 A: 分享链接 (https://cloud.189.cn/t/xxxx)
			if matches := reShareLink.FindStringSubmatch(resourceStr); len(matches) > 1 {
				shareCode := matches[1]

				mp = &models.MountPoint{
					Name:      "分享导入_" + shareCode, // 初始名
					TokenName: shareCode,           // 存储 ShareCode
					Password:  currentAccessCode,   // 存储访问码
					OsType:    protocolSubscribeShare, // 使用 handler.go 定义的常量
				}

			} else if reFolderID.MatchString(resourceStr) {
				// 情况 B: 纯数字 (文件夹 ID)]
				fileId, err := strconv.ParseInt(resourceStr, 10, 64)
				if err != nil {
					ctx.GetContext().Warn("ID转换失败", zap.String("line", line))
					failCount++
					continue
				}

				mp = &models.MountPoint{
					Name:   "文件夹导入_" + resourceStr, // 初始名
					FileId: fileId,
					OsType: protocolSubscribe, // 使用 handler.go 定义的常量
				}

			} else {
				// 无法识别
				ctx.GetContext().Warn("无法识别的行格式", zap.String("line", line))
				failCount++
				continue
			}

			// 3. 应用公共配置 (前端传来的设置)
			mp.TokenId = req.CloudToken
			mp.EnableAutoRefresh = req.EnableAutoRefresh
			mp.RefreshInterval = req.RefreshInterval

			// 4. 保存到数据库
			if _, err := h.mountPointService.Create(ctx.GetContext(), mp); err != nil {
				// 记录错误但不中断
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
