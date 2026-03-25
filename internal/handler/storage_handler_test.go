package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func newTestServer() *Server {
	return &Server{
		lgr: zap.NewNop(),
	}
}

type handlerCase struct {
	name       string
	method     string
	target     string
	body       string
	call       func(s *Server, w http.ResponseWriter, r *http.Request)
	wantStatus int
	wantBody   string
}

func runHandlerCase(t *testing.T, tc handlerCase) {
	t.Helper()

	s := newTestServer()

	var bodyReader *strings.Reader
	if tc.body == "" {
		bodyReader = strings.NewReader("")
	} else {
		bodyReader = strings.NewReader(tc.body)
	}

	req := httptest.NewRequest(tc.method, tc.target, bodyReader)
	if tc.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	tc.call(s, rec, req)

	if rec.Code != tc.wantStatus {
		t.Fatalf("want status %d, got %d, body=%q", tc.wantStatus, rec.Code, rec.Body.String())
	}

	if tc.wantBody != "" && !strings.Contains(rec.Body.String(), tc.wantBody) {
		t.Fatalf("want body to contain %q, got %q", tc.wantBody, rec.Body.String())
	}
}

func TestStorageHandlers_Validation(t *testing.T) {
	tests := []handlerCase{
		// AuthInfo
		{
			name:       "CreateAuthInfo bad json",
			method:     http.MethodPost,
			target:     "/api/auth",
			body:       `{"site":"test"`,
			call:       (*Server).CreateAuthInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid request",
		},
		{
			name:       "GetAuthInfo missing id",
			method:     http.MethodGet,
			target:     "/api/auth",
			call:       (*Server).GetAuthInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Нет параметра id",
		},
		{
			name:       "UpdateAuthInfo bad json",
			method:     http.MethodPut,
			target:     "/api/auth",
			body:       `{"id":"123"`,
			call:       (*Server).UpdateAuthInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "невалидный запрос",
		},
		{
			name:       "DeleteAuthInfo missing id",
			method:     http.MethodDelete,
			target:     "/api/auth?version=1",
			call:       (*Server).DeleteAuthInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Нет параметра id",
		},
		{
			name:       "DeleteAuthInfo missing version current behavior",
			method:     http.MethodDelete,
			target:     "/api/auth?id=123",
			call:       (*Server).DeleteAuthInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Нет параметра version",
		},
		{
			name:       "DeleteAuthInfo bad version",
			method:     http.MethodDelete,
			target:     "/api/auth?id=123&version=abc",
			call:       (*Server).DeleteAuthInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Некорректный параметр version",
		},

		// TextInfo
		{
			name:       "CreateTextInfo bad json",
			method:     http.MethodPost,
			target:     "/api/text",
			body:       `{"text":"hello"`,
			call:       (*Server).CreateTextInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Невалидный запрос",
		},
		{
			name:       "GetTextInfo missing id",
			method:     http.MethodGet,
			target:     "/api/text",
			call:       (*Server).GetTextInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Нет параметра id",
		},
		{
			name:       "UpdateTextInfo bad json",
			method:     http.MethodPut,
			target:     "/api/text",
			body:       `{"id":"123"`,
			call:       (*Server).UpdateTextInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Невалидный запрос",
		},
		{
			name:       "DeleteTextInfo missing id",
			method:     http.MethodDelete,
			target:     "/api/text?version=1",
			call:       (*Server).DeleteTextInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Нет параметра id",
		},
		{
			name:       "DeleteTextInfo missing version current behavior",
			method:     http.MethodDelete,
			target:     "/api/text?id=123",
			call:       (*Server).DeleteTextInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Нет параметра version",
		},
		{
			name:       "DeleteTextInfo bad version",
			method:     http.MethodDelete,
			target:     "/api/text?id=123&version=abc",
			call:       (*Server).DeleteTextInfo,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Некорректный параметр version",
		},

		// BankCard
		{
			name:       "CreateBankCard bad json",
			method:     http.MethodPost,
			target:     "/api/bankcard",
			body:       `{"card_number":"1111"`,
			call:       (*Server).CreateBankCard,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid request",
		},
		{
			name:       "GetBankCard missing id",
			method:     http.MethodGet,
			target:     "/api/bankcard",
			call:       (*Server).GetBankCard,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Нет параметра id",
		},
		{
			name:       "UpdateBankCard bad json",
			method:     http.MethodPut,
			target:     "/api/bankcard",
			body:       `{"id":"123"`,
			call:       (*Server).UpdateBankCard,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid request",
		},
		{
			name:       "DeleteBankCard missing id",
			method:     http.MethodDelete,
			target:     "/api/bankcard?version=1",
			call:       (*Server).DeleteBankCard,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Нет параметра id",
		},
		{
			name:       "DeleteBankCard missing version current behavior",
			method:     http.MethodDelete,
			target:     "/api/bankcard?id=123",
			call:       (*Server).DeleteBankCard,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Нет параметра version",
		},
		{
			name:       "DeleteBankCard bad version",
			method:     http.MethodDelete,
			target:     "/api/bankcard?id=123&version=abc",
			call:       (*Server).DeleteBankCard,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Некорректный параметр version",
		},

		// File
		{
			name:       "CreateFile bad json",
			method:     http.MethodPost,
			target:     "/api/file",
			body:       `{"path":"/tmp/a.txt"`,
			call:       (*Server).CreateFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "невалидный json",
		},
		{
			name:       "CreateFile empty path",
			method:     http.MethodPost,
			target:     "/api/file",
			body:       `{}`,
			call:       (*Server).CreateFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "не задан путь к файлу",
		},
		{
			name:       "UpdateFile bad json",
			method:     http.MethodPut,
			target:     "/api/file",
			body:       `{"path":"/tmp/a.txt"`,
			call:       (*Server).UpdateFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "невалидный json",
		},
		{
			name:       "UpdateFile empty path",
			method:     http.MethodPut,
			target:     "/api/file",
			body:       `{"version":1}`,
			call:       (*Server).UpdateFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "не задан путь к файлу",
		},
		{
			name:       "UpdateFile zero version",
			method:     http.MethodPut,
			target:     "/api/file",
			body:       `{"path":"/tmp/a.txt","version":0}`,
			call:       (*Server).UpdateFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "не задана версия",
		},
		{
			name:       "GetFile missing path",
			method:     http.MethodGet,
			target:     "/api/file?file_id=f1&version=1",
			call:       (*Server).GetFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "не указан путь",
		},
		{
			name:       "GetFile missing file id",
			method:     http.MethodGet,
			target:     "/api/file?path=/tmp",
			call:       (*Server).GetFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "не указан id файла",
		},
		{
			name:       "GetFile bad version",
			method:     http.MethodGet,
			target:     "/api/file?path=/tmp&file_id=f1&version=abc",
			call:       (*Server).GetFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "невалидная версия",
		},
		{
			name:       "DeleteFile missing file_id",
			method:     http.MethodDelete,
			target:     "/api/file?version=1",
			call:       (*Server).DeleteFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "не указан id файла",
		},
		{
			name:       "DeleteFile missing version",
			method:     http.MethodDelete,
			target:     "/api/file?file_id=f1",
			call:       (*Server).DeleteFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "не указана версия файла",
		},
		{
			name:       "DeleteFile bad version",
			method:     http.MethodDelete,
			target:     "/api/file?file_id=f1&version=abc",
			call:       (*Server).DeleteFile,
			wantStatus: http.StatusBadRequest,
			wantBody:   "невалидная версия",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runHandlerCase(t, tc)
		})
	}
}
