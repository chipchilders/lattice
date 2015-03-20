package receptor_json_runner

import (
    "encoding/json"
    "github.com/cloudfoundry-incubator/receptor"
)

//go:generate counterfeiter -o fake_receptor_json_runner/fake_receptor_json_runner.go . ReceptorJsonRunner
type ReceptorJsonRunner interface {
	CreateAppFromJson(json string) error
}

type receptorJsonRunner struct {
    receptorClient receptor.Client
    createRequest receptor.DesiredLRPCreateRequest
}

func New(receptorClient receptor.Client, createRequest receptor.DesiredLRPCreateRequest) ReceptorJsonRunner {
    return &receptorJsonRunner{receptorClient, createRequest}
}


func (r *receptorJsonRunner) CreateAppFromJson(paramsInJson string) error {

    request := r.createRequest
    err := json.Unmarshal([]byte(paramsInJson), &request)
    if err != nil {
        return err
    }

// check if app already running
//    if exists, err := r.appRunner.desiredLRPExists(request.Name); err != nil {
//        return err
//    } else if exists {
//        return newExistingAppError(request.Name)
//    }

    if err := r.receptorClient.UpsertDomain(request.Domain, 0); err != nil {
        return err
    }

    err = r.receptorClient.CreateDesiredLRP(request)
    if err != nil {
        return err
    }
//
//    return appRunner.desireLrp(request)
//}

	return nil
}
