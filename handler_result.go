package main

type HandlerResult struct {
	Reply      string
	Error      error
	BroadCast  bool
	RemindMode bool
}

func MakeHandlerResultSuccess(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, false, false}
}

func MakeHandlerResultBroadcast(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, true, false}
}

func MakeHandlerResultRemind(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, true, true}
}

func MakeHandlerResultError(e error) *HandlerResult {
	return &HandlerResult{"", e, false, false}
}
