package main

type HandlerResult struct {
	Reply       string
	Error       error
	BroadCast   bool
	RemindMode  bool
	ContactMode bool
}

func MakeHandlerResultSuccess(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, false, false, false}
}

func MakeHandlerResultBroadcast(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, true, false, false}
}

func MakeHandlerResultRemind(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, true, true, false}
}

func MakeHandlerResultContact(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, false, false, true}
}

func MakeHandlerResultError(e error) *HandlerResult {
	return &HandlerResult{"", e, false, false, false}
}
