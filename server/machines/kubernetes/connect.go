package kubernetes

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/layer5io/meshery/server/machines"
	"github.com/layer5io/meshery/server/models"
	"github.com/layer5io/meshery/server/models/connections"
	"github.com/layer5io/meshkit/models/events"
)

type ConnectAction struct {}

// Execute On Entry and Exit should not return next eventtype i suppose, look again.
func(ca *ConnectAction) ExecuteOnEntry(ctx context.Context, machineCtx interface{}) (machines.EventType, *events.Event, error) {
	fmt.Println("inside connect OnEntry.")
	eventBuilder := events.NewEvent().ActedUpon(uuid.Nil).WithCategory("connection").WithAction("register").FromSystem(uuid.Nil).FromUser(uuid.Nil) // pass userID and systemID in acted upon first pass user id if we can get context then update with connection Id
	machinectx, err := GetMachineCtx(machineCtx, eventBuilder)
	if err != nil {
		return machines.NoOp, eventBuilder.Build(), err
	}
	
	err = AssignClientSetToContext(machinectx, eventBuilder)
	if err != nil {
		return machines.NoOp, eventBuilder.Build(), err
	}
	err = machinectx.K8sContext.PingTest()
	fmt.Println("inside connect line 30 after ping test", err)
	if err != nil {
		// machinectx.log.Error(err)
		fmt.Println("inside connect line 30 after ping test err block", err)
		// peform error handling and event publishing
		return machines.NotFound, nil, err
	}

	token, _ := ctx.Value(models.TokenCtxKey).(string)
	connection, statusCode, err := machinectx.Provider.UpdateConnectionStatusByID(token, uuid.FromStringOrNil(machinectx.K8sContext.ConnectionID), connections.CONNECTED)
	// peform error handling and event publishing
	if err != nil {
		return machines.NoOp, eventBuilder.Build(), err
	}

	fmt.Println("connection inside connect.go updated connection status", connection, statusCode)

	return machines.NoOp, eventBuilder.Build(), nil

}
func(ca *ConnectAction) Execute(ctx context.Context, machineCtx interface{}) (machines.EventType, *events.Event, error) {
	fmt.Println("inside connect Execute mch")
	eventBuilder := events.NewEvent().ActedUpon(uuid.Nil).WithCategory("connection").WithAction("register").FromSystem(uuid.Nil).FromUser(uuid.Nil) // pass userID and systemID in acted upon first pass user id if we can get context then update with connection Id
	machinectx, err := GetMachineCtx(machineCtx, eventBuilder)
	if err != nil {
		return machines.NoOp, eventBuilder.Build(), err
	}

	k8sContexts := []models.K8sContext{machinectx.K8sContext}
	machinectx.MesheryCtrlsHelper.UpdateCtxControllerHandlers(k8sContexts).
	UpdateOperatorsStatusMap(machinectx.OperatorTracker).DeployUndeployedOperators(machinectx.OperatorTracker)
	
	ctx = context.WithValue(ctx, models.MesheryControllerHandlersKey, machinectx. MesheryCtrlsHelper.GetControllerHandlersForEachContext())

	return machines.NoOp, eventBuilder.Build(), nil
}

func(ca *ConnectAction) ExecuteOnExit(ctx context.Context, machineCtx interface{}) (machines.EventType, *events.Event, error) {
	return machines.NoOp, nil, nil
}
