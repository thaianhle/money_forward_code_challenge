package httpresponse

import "net/http"

type Response struct {
	Code          int    `json:"code"`
	ErrCodeString string `json:"err_code_string"`
	Data          any    `json:"data"`
}

func (r *Response) resetBeforeTransform() {
	r.Code = 0
	r.ErrCodeString = ""
	r.Data = nil
}
func (r *Response) ErrString() *string {
	return &r.ErrCodeString
}

func (r *Response) constructErrMessage(errString string) *Response {
	r.ErrCodeString = errString
	return r
}

func (r *Response) constructDataMessage(dataMessage any) *Response {
	r.Data = dataMessage
	return r
}

func (r *Response) TransformToInternalServerError(errString string) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusInternalServerError
	return r.constructErrMessage(errString)
}

func (r *Response) TransformToBadRequest(errString string) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusBadRequest
	return r.constructErrMessage(errString)
}

func (r *Response) TransformToNotFound(errString string) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusNotFound
	return r.constructErrMessage(errString)
}

func (r *Response) TransformToCreatedSuccess(dataMessage any) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusCreated
	return r.constructDataMessage(dataMessage)
}

func (r *Response) TransformToUpdatedSuccess(dataMessage any) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusAccepted
	return r.constructDataMessage(dataMessage)
}

func (r *Response) TransformToDeletedSuccess(dataMessage any) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusAccepted
	return r.constructDataMessage(dataMessage)
}

func (r *Response) TransformToConflictUniqueResourceError(errString string) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusConflict
	return r.constructErrMessage(errString)
}

func (r *Response) TransformToSuccessOk(data any) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusOK
	return r.constructDataMessage(data)
}
