package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/konkovaanna23/gophkeeper/internal/handler/helper"
	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
	"go.uber.org/zap"
)

type filePostRequest struct {
	Path string            `json:"path"`
	Meta map[string]string `json:"meta"`
}

type filePutRequest struct {
	Path    string            `json:"path"`
	Meta    map[string]string `json:"meta"`
	Version int64             `json:"version"`
}

type fileResponse struct {
	FileID    string            `json:"file_id,omitempty"`
	Version   int64             `json:"version,omitempty"`
	Path      string            `json:"path,omitempty"`
	Filename  string            `json:"filename,omitempty"`
	SizeBytes int64             `json:"size_bytes,omitempty"`
	Checksum  string            `json:"checksum,omitempty"`
	Status    string            `json:"status,omitempty"`
	Meta      map[string]string `json:"meta"`
}

// CreateAuthInfo godoc
// @Summary      Сохранить учётные данные
// @Description  Сохраняет логин и пароль
// @Tags         credentials
// @Accept       json
// @Produce      json
// @Param        input body pb.AuthInfo true "Данные: login, password, meta"
// @Success      200 {object} dto.CreateResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/auth-info [post]
func (s *Server) CreateAuthInfo(w http.ResponseWriter, r *http.Request) {
	var req pb.AuthInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	resp, err := s.storageClient.CreateAuthInfo(r.Context(), &req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// GetAuthInfo godoc
// @Summary      Получить учётные данные
// @Description  Возвращает сохранённые логин/пароль по ID
// @Tags         credentials
// @Produce      json
// @Param        id query string true "ID записи"
// @Success      200 {object} pb.StoredAuthInfo
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/auth-info [get]
func (s *Server) GetAuthInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Нет параметра id", http.StatusBadRequest)
		return
	}

	req := &pb.GetInfoRequest{Id: id}
	resp, err := s.storageClient.GetAuthInfo(r.Context(), req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// UpdateAuthInfo godoc
// @Summary      Обновить учётные данные
// @Description  Обновляет запись с проверкой версии
// @Tags         credentials
// @Accept       json
// @Produce      json
// @Param        input body pb.UpdateAuthInfoRequest true "ID, data, version"
// @Success      200 {object} dto.UpdateResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse "Конфликт версий"
// @Failure      500 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/auth-info [put]
func (s *Server) UpdateAuthInfo(w http.ResponseWriter, r *http.Request) {
	var req pb.UpdateAuthInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.lgr.Error("невалидный запрос", zap.Error(err))
		http.Error(w, "невалидный запрос", http.StatusBadRequest)
		return
	}

	resp, err := s.storageClient.UpdateAuthInfo(r.Context(), &req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// DeleteAuthInfo godoc
// @Summary      Удалить учётные данные
// @Description  Удаляет запись по ID и версии
// @Tags         credentials
// @Produce      json
// @Param        id query string true "ID записи"
// @Param        version query int true "Текущая версия"
// @Success      200 {object} pb.DeleteInfoResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/auth-info [delete]
func (s *Server) DeleteAuthInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Нет параметра id", http.StatusBadRequest)
		return
	}

	version := r.URL.Query().Get("version")
	if version == "" {
		http.Error(w, "Нет параметра version", http.StatusBadRequest)
		return
	}

	versionInt, err := strconv.Atoi(version)
	if err != nil {
		http.Error(w, "Некорректный параметр version", http.StatusBadRequest)
		return
	}

	req := &pb.DeleteInfoRequest{Id: id, Version: int64(versionInt)}

	resp, err := s.storageClient.DeleteAuthInfo(r.Context(), req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// CreateTextInfo godoc
// @Summary      Сохранить текстовую заметку
// @Description  Сохраняет произвольный текст
// @Tags         text
// @Accept       json
// @Produce      json
// @Param        input body pb.TextInfo true "Текст и метаданные"
// @Success      200 {object} dto.CreateResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/text-info [post]
func (s *Server) CreateTextInfo(w http.ResponseWriter, r *http.Request) {
	var req pb.TextInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Невалидный запрос", http.StatusBadRequest)
		return
	}

	resp, err := s.storageClient.CreateTextInfo(r.Context(), &req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// GetTextInfo godoc
// @Summary      Получить текстовую заметку
// @Description  Возвращает текст по ID
// @Tags         text
// @Produce      json
// @Param        id query string true "ID записи"
// @Success      200 {object} pb.StoredTextInfo
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/text-info [get]
func (s *Server) GetTextInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Нет параметра id", http.StatusBadRequest)
		return
	}

	req := &pb.GetInfoRequest{Id: id}
	resp, err := s.storageClient.GetTextInfo(r.Context(), req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// UpdateTextInfo godoc
// @Summary      Обновить текстовую заметку
// @Description  Обновляет текст с контролем версий
// @Tags         text
// @Accept       json
// @Produce      json
// @Param        input body pb.UpdateTextInfoRequest true "ID, data, version"
// @Success      200 {object} dto.UpdateResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/text-info [put]
func (s *Server) UpdateTextInfo(w http.ResponseWriter, r *http.Request) {
	var req pb.UpdateTextInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Невалидный запрос", http.StatusBadRequest)
		return
	}

	resp, err := s.storageClient.UpdateTextInfo(r.Context(), &req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// DeleteTextInfo godoc
// @Summary      Удалить текстовую заметку
// @Description  Удаляет запись по ID и версии
// @Tags         text
// @Produce      json
// @Param        id query string true "ID записи"
// @Param        version query int true "Версия"
// @Success      200 {object} pb.DeleteInfoResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/text-info [delete]
func (s *Server) DeleteTextInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Нет параметра id", http.StatusBadRequest)
		return
	}

	version := r.URL.Query().Get("version")
	if version == "" {
		http.Error(w, "Нет параметра version", http.StatusBadRequest)
		return
	}

	versionInt, err := strconv.Atoi(version)
	if err != nil {
		http.Error(w, "Некорректный параметр version", http.StatusBadRequest)
		return
	}

	req := &pb.DeleteInfoRequest{Id: id, Version: int64(versionInt)}

	resp, err := s.storageClient.DeleteTextInfo(r.Context(), req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// CreateBankCard godoc
// @Summary      Сохранить данные банковской карты
// @Description  Сохраняет номер, срок, имя держателя
// @Tags         bank-cards
// @Accept       json
// @Produce      json
// @Param        input body pb.BankCardDetails true "Данные карты"
// @Success      200 {object} dto.CreateResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/bank-card [post]
func (s *Server) CreateBankCard(w http.ResponseWriter, r *http.Request) {
	var req pb.BankCardDetails
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	resp, err := s.storageClient.CreateBankCardDetails(r.Context(), &req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// GetBankCard godoc
// @Summary      Получить данные банковской карты
// @Description  Возвращает номер, срок, имя держателя
// @Tags         bank-cards
// @Produce      json
// @Param        id query string true "ID записи"
// @Success      200 {object} pb.StoredTextInfo
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/bank-card [get]
func (s *Server) GetBankCard(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Нет параметра id", http.StatusBadRequest)
		return
	}

	req := &pb.GetInfoRequest{Id: id}
	resp, err := s.storageClient.GetBankCardDetails(r.Context(), req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// UpdateBankCard godoc
// @Summary      Обновить данные банковской карты
// @Description  Обновляет номер, срок, имя держателя
// @Tags         bank-cards
// @Accept       json
// @Produce      json
// @Param        input body pb.UpdateBankCardDetailsRequest true "ID, data, version"
// @Success      200 {object} dto.UpdateResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/bank-card [put]
func (s *Server) UpdateBankCard(w http.ResponseWriter, r *http.Request) {
	var req pb.UpdateBankCardDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	resp, err := s.storageClient.UpdateBankCardDetails(r.Context(), &req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// DeleteBankCard godoc
// @Summary      Удалить данные банковской карты
// @Description  Удаляет запись по ID и версии
// @Tags         bank-cards
// @Produce      json
// @Param        id query string true "ID записи"
// @Param        version query int true "Версия"
// @Success      200 {object} pb.DeleteInfoResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/bank-card [delete]
func (s *Server) DeleteBankCard(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Нет параметра id", http.StatusBadRequest)
		return
	}

	version := r.URL.Query().Get("version")
	if version == "" {
		http.Error(w, "Нет параметра version", http.StatusBadRequest)
		return
	}

	versionInt, err := strconv.Atoi(version)
	if err != nil {
		http.Error(w, "Некорректный параметр version", http.StatusBadRequest)
		return
	}

	req := &pb.DeleteInfoRequest{Id: id, Version: int64(versionInt)}

	resp, err := s.storageClient.DeleteBankCardDetails(r.Context(), req)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// CreateFile godoc
// @Summary      Загрузить файл
// @Description  Начинает загрузку файла, отправляет чанки, финализирует
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Файл"
// @Param        meta formData string false "Метаданные (JSON)" format(json)
// @Success      200 {object} pb.CommitUploadResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/file [post]
func (s *Server) CreateFile(w http.ResponseWriter, r *http.Request) {
	var req filePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.lgr.Error("невалидный json", zap.Error(err))
		http.Error(w, "невалидный json", http.StatusBadRequest)
		return
	}
	if req.Path == "" {
		http.Error(w, "не задан путь к файлу", http.StatusBadRequest)
		return
	}

	resp, err := s.storageClient.CreateFileFromPath(r.Context(), req.Path, req.Meta)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, fileResponse{
		FileID:  resp.FileId,
		Version: resp.NewVersion,
		Status:  "uploaded",
	})
}

// UpdateFile godoc
// @Summary      Обновить файл
// @Description  Обновляет существующий файл с контролем версий
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Новый файл"
// @Param        id query string true "ID файла"
// @Param        version query int true "Текущая версия"
// @Param        meta formData string false "Метаданные" format(json)
// @Success      200 {object} pb.CommitUploadResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/file [put]
func (s *Server) UpdateFile(w http.ResponseWriter, r *http.Request) {
	var req filePutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.lgr.Error("невалидный json", zap.Error(err))
		http.Error(w, "невалидный json", http.StatusBadRequest)
		return
	}
	if req.Path == "" {
		http.Error(w, "не задан путь к файлу", http.StatusBadRequest)
		return
	}
	if req.Version == 0 {
		http.Error(w, "не задана версия", http.StatusBadRequest)
		return
	}

	resp, err := s.storageClient.UpdateFileFromPath(r.Context(), req.Path, req.Meta, req.Version)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, fileResponse{
		FileID:  resp.FileId,
		Version: resp.NewVersion,
		Status:  "updated",
	})
}

// GetFile godoc
// @Summary      Скачать файл
// @Description  Возвращает файл по ID
// @Tags         files
// @Produce      application/octet-stream
// @Param        id query string true "ID файла"
// @Param        version query int false "Версия (0 = последняя)" default(0)
// @Success      200 {file} file "Файл для скачивания"
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/file [get]
func (s *Server) GetFile(w http.ResponseWriter, r *http.Request) {
	dirPath := r.URL.Query().Get("path")
	fileid := r.URL.Query().Get("file_id")
	versionStr := r.URL.Query().Get("version")

	if dirPath == "" {
		http.Error(w, "не указан путь", http.StatusBadRequest)
		return
	}
	if fileid == "" {
		http.Error(w, "не указан id файла", http.StatusBadRequest)
		return
	}

	var version int64
	var err error
	if versionStr != "" {
		version, err = strconv.ParseInt(versionStr, 10, 64)
		if err != nil {
			http.Error(w, "невалидная версия", http.StatusBadRequest)
			return
		}
	}

	metaResp, fullPath, err := s.storageClient.DownloadFileToPath(r.Context(), dirPath, fileid, version)
	if err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, fileResponse{
		FileID:    metaResp.File.FileId,
		Version:   metaResp.File.CurrentVersion,
		Path:      fullPath,
		Filename:  metaResp.File.Filename,
		SizeBytes: metaResp.File.SizeBytes,
		Checksum:  metaResp.File.Checksum,
		Meta:      metaResp.File.Meta,
		Status:    "downloaded",
	})
}

// DeleteFile godoc
// @Summary      Удалить файл
// @Description  Удаляет файл по ID и версии
// @Tags         files
// @Produce      json
// @Param        id query string true "ID файла"
// @Param        version query int true "Текущая версия"
// @Success      200 {object} pb.DeleteInfoResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /storage/file [delete]
func (s *Server) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileid := r.URL.Query().Get("file_id")
	versionStr := r.URL.Query().Get("version")

	if fileid == "" {
		http.Error(w, "не указан id файла", http.StatusBadRequest)
		return
	}
	if versionStr == "" {
		http.Error(w, "не указана версия файла", http.StatusBadRequest)
		return
	}

	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		http.Error(w, "невалидная версия", http.StatusBadRequest)
		return
	}

	if err := s.storageClient.DeleteFile(r.Context(), fileid, version); err != nil {
		s.lgr.Error("ошибка grpc запроса", zap.Error(err))
		helper.WriteGRPCError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, fileResponse{
		FileID:  fileid,
		Version: version,
		Status:  "deleted",
	})
}
