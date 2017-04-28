package goodies

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CommandRequest Command like class to produce requests to GoodiesCommandProcessor
// Name would be the same as the method name exposed by goodies.Provider interface
// Parameters would be the method parameters respectively
// (ttl would be sent in seconds as string or as ExpireDefault/ExpireNever)
type CommandRequest struct {
	Name       string
	Parameters []string
}

// CommandResponse Command like class that is returned as a result of command execution
// Success will be set to true in case if no error happened when processing the command
// Result will contain command result value (lists will be serialised as comma separated)
// Err will contain typed error from the list of errors exposed by goodies package (see Err* types)
type CommandResponse struct {
	Success bool
	Result  string
	Err     error
}

func NewCommandResponseFromError(err error) CommandResponse {
	//TODO: add runtime check for known errors
	return CommandResponse{false, "", err}
}

func NewCommandResponseFromResult(res string) CommandResponse {
	return CommandResponse{true, res, nil}
}

// NewGoodiesCommandsProcessor Creates a generic command processor for goodies provider
func NewGoodiesCommandsProcessor(storage Provider) CommandProcessor {
	gcp := goodiesCommandProcessor{storage, make(map[string]func(command CommandRequest, storage Provider) CommandResponse, 1)}
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

// goodiesCommandProcessor Generic command processor class
// helping to wrap command processing to the storage and back
// mediates commands to provider
type goodiesCommandProcessor struct {
	storage         Provider
	commandHandlers map[string]func(command CommandRequest, storage Provider) CommandResponse
}

func (gcp *goodiesCommandProcessor) addCommandHandler(
	name string,
	handler func(command CommandRequest, storage Provider) CommandResponse) {

	gcp.commandHandlers[name] = handler
}

func (gcp *goodiesCommandProcessor) Process(req CommandRequest) CommandResponse {
	handler, ok := gcp.commandHandlers[req.Name]
	if !ok {
		return NewCommandResponseFromError(ErrUnknownCommand{req.Name})
	}
	resp := handler(req, gcp.storage)
	return resp
}

func setCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 3 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"Set command is expected to have 3 arguments (key, value, ttl)"})
	}
	ttl, err := parseTTL(command.Parameters[2])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	storage.Set(command.Parameters[0], command.Parameters[1], ttl)
	return NewCommandResponseFromResult(command.Parameters[1])
}

func getCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 1 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"Get command is expected to have 1 argument (key)"})
	}
	val, err := storage.Get(command.Parameters[0])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult(val)
}

func updateCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 3 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"Update command is expected to have 3 arguments (key, value, ttl)"})
	}
	ttl, err := parseTTL(command.Parameters[2])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	err = storage.Update(command.Parameters[0], command.Parameters[1], ttl)
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult("")
}

func removeCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 1 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"Remove command is expected to have 1 argument (key)"})
	}
	storage.Remove(command.Parameters[0])
	return NewCommandResponseFromResult("")
}

func keysCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 0 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"Keys command is expected to have 0 arguments"})
	}
	val, _ := storage.Keys()
	return NewCommandResponseFromResult(strings.Trim(strings.Join(strings.Fields(fmt.Sprint(val)), ":"), "[]"))
}

func listPushCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 2 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"ListPush command is expected to have 2 arguments (key, value)"})
	}
	err := storage.ListPush(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult("")
}

func listLenCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 1 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"ListLen command is expected to have 1 argument (key)"})
	}
	val, err := storage.ListLen(command.Parameters[0])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult(strconv.Itoa(val))
}

func listRemoveIndexCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 2 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"ListRemoveIndex command is expected to have 2 arguments (key, index(INT))"})
	}
	i, err := strconv.Atoi(command.Parameters[1])
	if err != nil {
		return NewCommandResponseFromError(
			ErrCommandArgumentsMismatch{"ListRemoveIndex command expects to receive index (2nd argument ) as integer "})
	}
	err = storage.ListRemoveIndex(command.Parameters[0], i)
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult("")
}

func listRemoveValueCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 2 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"ListRemoveValue command is expected to have 2 arguments (key, index(INT))"})
	}
	err := storage.ListRemoveValue(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult("")
}

func listGetByIndexCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 2 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"ListGetByIndex command is expected to have 2 arguments (key, index(INT))"})
	}
	i, err := strconv.Atoi(command.Parameters[1])
	if err != nil {
		return NewCommandResponseFromError(
			ErrCommandArgumentsMismatch{"ListGetByIndex command expects to receive index (2nd argument ) as integer "})
	}
	val, err := storage.ListGetByIndex(command.Parameters[0], i)
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult(val)
}

func dictSetCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 3 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"DictSet command is expected to have 3 arguments (key, dictKey, value)"})
	}
	err := storage.DictSet(command.Parameters[0], command.Parameters[1], command.Parameters[2])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult("")
}

func dictGetCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 2 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"DictGet command is expected to have 2 arguments (key, dictKey)"})
	}
	val, err := storage.DictGet(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult(val)
}

func dictRemoveCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 2 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"DictRemove command is expected to have 2 arguments (key, dictKey)"})
	}
	err := storage.DictRemove(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult("")
}

func dictHasKeyCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 2 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"DictHasKey command is expected to have 2 arguments (key, dictKey)"})
	}
	yes, err := storage.DictHasKey(command.Parameters[0], command.Parameters[1])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	if yes {
		return NewCommandResponseFromResult("1")
	}
	return NewCommandResponseFromResult("0")
}

func setExpiryCommandHandler(command CommandRequest, storage Provider) CommandResponse {
	if len(command.Parameters) != 2 {
		return NewCommandResponseFromError(ErrCommandArgumentsMismatch{"SetExpiry command is expected to have 2 argument (key, ttl(INT SECONDS))"})
	}
	ttl, err := parseTTL(command.Parameters[1])
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	err = storage.SetExpiry(command.Parameters[0], ttl)
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return NewCommandResponseFromResult("")
}

func parseTTL(s string) (time.Duration, error) {
	if s == "-2" {
		return ExpireDefault, nil
	}
	if s == "-1" {
		return ExpireNever, nil
	}
	seconds, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, ErrCommandArgumentsMismatch{"Ttl parameter is of unexpected format. Should be integer (nanoseconds)"}
	}

	return time.Duration(seconds) * time.Second, nil
}
