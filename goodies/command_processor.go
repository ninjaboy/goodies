package goodies

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type GoodiesRequest struct {
	Name       string
	Parameters []string
}

type GoodiesResponse struct {
	Success bool
	Result  string
	Err     error
}

type GoodiesCommandProcessor struct {
	storage         Provider
	commandHandlers map[string]func(command GoodiesRequest, storage Provider) GoodiesResponse
}

func (gcp *GoodiesCommandProcessor) addCommandHandler(
	name string,
	handler func(command GoodiesRequest, storage Provider) GoodiesResponse) {

	gcp.commandHandlers[name] = handler
}

func (gcp *GoodiesCommandProcessor) HandleCommand(req GoodiesRequest) GoodiesResponse {
	defer func() {
		if r := recover(); r != nil {

		}
	}()
	handler, ok := gcp.commandHandlers[req.Name]
	if !ok {
		return createErrorResult(ErrUnknownCommand{req.Name})
	}
	return handler(req, gcp.storage)
}

func createErrorResult(err error) GoodiesResponse {
	//TODO: add runtime check for known errors
	return GoodiesResponse{false, "", err}
}

func createOkResult(res string) GoodiesResponse {
	return GoodiesResponse{true, res, nil}
}

// NewGoodiesCommandsProcessor Creates a generic command processor for goodies provider
func NewGoodiesCommandsProcessor(storage Provider) *GoodiesCommandProcessor {
	gcp := GoodiesCommandProcessor{storage, make(map[string]func(command GoodiesRequest, storage Provider) GoodiesResponse, 1)}
	gcp.addCommandHandler("Set", setCommandHandler)
	gcp.addCommandHandler("Get", getCommandHandler)
	gcp.addCommandHandler("Update", updateCommandHandler)
	gcp.addCommandHandler("Remove", removeCommandHandler)
	gcp.addCommandHandler("Keys", keysCommandHandler)
	gcp.addCommandHandler("ListPush", listPushCommandHandler)
	gcp.addCommandHandler("ListLen", listLenCommandHandler)
	gcp.addCommandHandler("ListGetByIndex", listGetByIndexCommandHandler)
	gcp.addCommandHandler("ListRemoveIndex", listRemoveIndexCommandHandler)
	gcp.addCommandHandler("ListRemoveValue", listRemoveValueCommandHandler)
	gcp.addCommandHandler("DictSet", dictSetCommandHandler)
	gcp.addCommandHandler("DictGet", dictGetCommandHandler)
	gcp.addCommandHandler("DictRemove", dictRemoveCommandHandler)
	gcp.addCommandHandler("DictHasKey", dictHasKeyCommandHandler)
	gcp.addCommandHandler("SetExpiry", setExpiryCommandHandler)
	return &gcp
}

func setCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 3 {
		return createErrorResult(ErrCommandArgumentsMismatch{"Set command is expected to have 3 arguments (key, value, ttl)"})
	}
	ttl, err := parseTTL(command.Parameters[2])
	if err != nil {
		return createErrorResult(err)
	}
	storage.Set(command.Parameters[0], command.Parameters[1], ttl)
	return createOkResult("")
}

func getCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 1 {
		return createErrorResult(ErrCommandArgumentsMismatch{"Get command is expected to have 1 argument (key)"})
	}
	val, err := storage.Get(command.Parameters[0])
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult(val)
}

func updateCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 3 {
		return createErrorResult(ErrCommandArgumentsMismatch{"Update command is expected to have 3 arguments (key, value, ttl)"})
	}
	ttl, err := parseTTL(command.Parameters[2])
	if err != nil {
		return createErrorResult(err)
	}
	err = storage.Update(command.Parameters[0], command.Parameters[1], ttl)
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult("")
}

func removeCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 1 {
		return createErrorResult(ErrCommandArgumentsMismatch{"Remove command is expected to have 1 argument (key)"})
	}
	storage.Remove(command.Parameters[0])
	return createOkResult("")
}

func keysCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 0 {
		return createErrorResult(ErrCommandArgumentsMismatch{"Keys command is expected to have 0 arguments"})
	}
	val := storage.Keys()
	return createOkResult(strings.Trim(strings.Join(strings.Fields(fmt.Sprint(val)), ":"), "[]"))
}

func listPushCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 2 {
		return createErrorResult(ErrCommandArgumentsMismatch{"ListPush command is expected to have 2 arguments (key, value)"})
	}
	err := storage.ListPush(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult("")
}

func listLenCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 1 {
		return createErrorResult(ErrCommandArgumentsMismatch{"ListLen command is expected to have 1 argument (key)"})
	}
	val, err := storage.ListLen(command.Parameters[0])
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult(string(val))
}

func listRemoveIndexCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 2 {
		return createErrorResult(ErrCommandArgumentsMismatch{"ListRemoveIndex command is expected to have 2 arguments (key, index(INT))"})
	}
	i, err := strconv.Atoi(command.Parameters[1])
	if err != nil {
		return createErrorResult(
			ErrCommandArgumentsMismatch{"ListRemoveIndex command expects to receive index (2nd argument ) as integer "})
	}
	err = storage.ListRemoveIndex(command.Parameters[0], i)
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult("")
}

func listRemoveValueCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 2 {
		return createErrorResult(ErrCommandArgumentsMismatch{"ListRemoveValue command is expected to have 2 arguments (key, index(INT))"})
	}
	err := storage.ListRemoveValue(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult("")
}

func listGetByIndexCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 2 {
		return createErrorResult(ErrCommandArgumentsMismatch{"ListGetByIndex command is expected to have 2 arguments (key, index(INT))"})
	}
	i, err := strconv.Atoi(command.Parameters[1])
	if err != nil {
		return createErrorResult(
			ErrCommandArgumentsMismatch{"ListGetByIndex command expects to receive index (2nd argument ) as integer "})
	}
	val, err := storage.ListGetByIndex(command.Parameters[0], i)
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult(val)
}

func dictSetCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 3 {
		return createErrorResult(ErrCommandArgumentsMismatch{"DictSet command is expected to have 3 arguments (key, dictKey, value)"})
	}
	err := storage.DictSet(command.Parameters[0], command.Parameters[1], command.Parameters[2])
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult("")
}

func dictGetCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 2 {
		return createErrorResult(ErrCommandArgumentsMismatch{"DictGet command is expected to have 2 arguments (key, dictKey)"})
	}
	val, err := storage.DictGet(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult(val)
}

func dictRemoveCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 2 {
		return createErrorResult(ErrCommandArgumentsMismatch{"DictRemove command is expected to have 2 arguments (key, dictKey)"})
	}
	err := storage.DictRemove(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult("")
}

func dictHasKeyCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 2 {
		return createErrorResult(ErrCommandArgumentsMismatch{"DictHasKey command is expected to have 2 arguments (key, dictKey)"})
	}
	yes, err := storage.DictHasKey(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return createErrorResult(err)
	}
	if yes {
		return createOkResult("1")
	}
	return createOkResult("0")
}

func setExpiryCommandHandler(command GoodiesRequest, storage Provider) GoodiesResponse {
	if len(command.Parameters) != 2 {
		return createErrorResult(ErrCommandArgumentsMismatch{"SetExpiry command is expected to have 2 argument (key, ttl(INT SECONDS))"})
	}
	ttl, err := parseTTL(command.Parameters[1])
	if err != nil {
		return createErrorResult(err)
	}
	err = storage.SetExpiry(command.Parameters[0], ttl)
	if err != nil {
		return createErrorResult(err)
	}
	return createOkResult("")
}

func parseTTL(s string) (time.Duration, error) {
	nanoseconds, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, ErrCommandArgumentsMismatch{"Ttl parameter is of unexpected format. Should be integer (nanoseconds)"}
	}
	return time.Duration(nanoseconds) * time.Nanosecond, nil
}
