// Package helper предоставляет утилиты для HTTP-обработчиков, связанных с gRPC.
//
// Основные функции:
//   - WriteJSON: отправка JSON-ответа с заданным статусом.
//   - WriteGRPCError: преобразование gRPC-ошибки в HTTP-ответ с соответствующим статусом.
//   - grpcCodeToHTTPStatus: маппинг кодов gRPC на HTTP-статусы.
//
// Используется для создания REST-like API поверх gRPC-сервисов (например, через gateway или адаптеры).
//
// Пример:
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    _, err := grpcClient.SomeMethod(r.Context(), req)
//	    if err != nil {
//	        helper.WriteGRPCError(w, err)
//	        return
//	    }
//	    helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
//	}
package helper

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WriteGRPCError преобразует gRPC-ошибку в HTTP-ответ.
//
// Если ошибка:
//   - является gRPC-статусом (status.Error), то используется её Code() для определения HTTP-статуса.
//   - не является gRPC-статусом, возвращается 500 Internal Server Error.
//
// Отправляет JSON вида:
//
//	{
//	  "ok": false,
//	  "error": "сообщение об ошибке"
//	}
//
// Параметры:
//   - w: http.ResponseWriter.
//   - err: ошибка от gRPC-вызова.
//
// Пример:
//
//	_, err := client.Login(ctx, req)
//	if err != nil {
//	    helper.WriteGRPCError(w, err) // например, 401 Unauthorized
//	    return
//	}
func WriteGRPCError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	WriteJSON(w, grpcCodeToHTTPStatus(st.Code()), map[string]any{
		"ok":    false,
		"error": st.Message(),
	})
}

func grpcCodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// WriteJSON отправляет JSON-ответ клиенту с указанным HTTP-статусом.
//
// Устанавливает заголовок:
//   - Content-Type: application/json
//
// Параметры:
//   - w: http.ResponseWriter.
//   - statusCode: HTTP-код (например, 200, 400, 500).
//   - v: данные для сериализации в JSON.
//
// Пример:
//
//	helper.WriteJSON(w, http.StatusOK, map[string]any{
//	    "user_id": "123",
//	    "active":  true,
//	})
func WriteJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}
