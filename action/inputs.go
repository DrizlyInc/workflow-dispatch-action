package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v37/github"
	"github.com/sethvargo/go-githubactions"
)

type inputs struct {
	appID              int64
	privateKey         *rsa.PrivateKey
	targetRepository   string
	targetOwner        string
	targetRef          string
	workflowFilename   string
	waitForCheck       bool
	waitTimeoutSeconds int64
	installationId     int64
	workflowInputs     map[string]interface{}
}

func parseInputs() (inputs, error) {

	appIDString, ok := os.LookupEnv("INPUT_APP_ID")
	if !ok {
		return inputs{}, errors.New("input 'app_id' not set")
	}
	appID, err := strconv.ParseInt(appIDString, 10, 64)
	if err != nil {
		return inputs{}, errors.New("input 'app_id' must be an integer")
	}

	installationId := int64(-1)
	installationIdString, ok := os.LookupEnv("APP_INSTALLATION_ID")
	if ok {
		installationId, err = strconv.ParseInt(installationIdString, 10, 64)
		if err != nil {
			return inputs{}, errors.New("APP_INSTALLATION_ID must be an integer")
		}
	}

	privateKeyString, ok := os.LookupEnv("INPUT_PRIVATE_KEY")
	if !ok {
		return inputs{}, errors.New("input 'private_key' not set")
	}
	block, _ := pem.Decode([]byte(privateKeyString))
	if block == nil {
		return inputs{}, errors.New("input 'private_key' not a PEM block")
	}
	if block.Type != "RSA PRIVATE KEY" {
		return inputs{}, fmt.Errorf("input 'private_key' PEM block not an RSA private key. It is a %v", block.Type)
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return inputs{}, fmt.Errorf("input 'private_key' RSA Private Key not formatted properly: %w", err)
	}

	targetRepository, ok := os.LookupEnv("INPUT_TARGET_REPOSITORY")
	if !ok {
		return inputs{}, errors.New("input 'target_repository' not set")
	}
	targetOwnerRepo := strings.Split(targetRepository, "/")
	if len(targetOwnerRepo) != 2 {
		return inputs{}, errors.New("input 'target_repository' not formatted as owner/repo-name")
	}
	targetOwner := targetOwnerRepo[0]
	targetRepository = targetOwnerRepo[1]

	targetRef, ok := os.LookupEnv("INPUT_TARGET_REF")
	if !ok {
		return inputs{}, errors.New("input 'target_ref' not set")
	}

	workflowFilename, ok := os.LookupEnv("INPUT_WORKFLOW_FILENAME")
	if !ok {
		return inputs{}, errors.New("input 'workflow_filename' not set")
	}

	wfcString := os.Getenv("INPUT_WAIT_FOR_CHECK")
	waitForCheck, err := strconv.ParseBool(wfcString)
	if err != nil {
		return inputs{}, fmt.Errorf("input 'wait_for_check' is not a boolean: %w", err)
	}

	waitTimeoutSecondsString, ok := os.LookupEnv("INPUT_WAIT_TIMEOUT_SECONDS")
	if !ok {
		return inputs{}, errors.New("input 'wait_timeout_seconds' not set")
	}
	waitTimeoutSeconds, err := strconv.ParseInt(waitTimeoutSecondsString, 10, 64)
	if err != nil {
		return inputs{}, errors.New("input 'wait_timeout_seconds' must be an integer")
	}

	workflowInputs := map[string]interface{}{}
	workflowInputsString, ok := os.LookupEnv("INPUT_WORKFLOW_INPUTS")
	if !ok {
		return inputs{}, errors.New("input 'workflow_inputs' not set")
	}
	err = json.Unmarshal([]byte(workflowInputsString), &workflowInputs)
	if err != nil {
		return inputs{}, fmt.Errorf("input 'workflow_inputs' is not json: %w", err)
	}

	return inputs{
		appID:              appID,
		privateKey:         privateKey,
		workflowFilename:   workflowFilename,
		targetRepository:   targetRepository,
		targetOwner:        targetOwner,
		targetRef:          targetRef,
		waitForCheck:       waitForCheck,
		waitTimeoutSeconds: waitTimeoutSeconds,
		workflowInputs:     workflowInputs,
		installationId:     installationId,
	}, nil
}

func addDefaultWorkflowInputs(inputs *inputs, githubVars githubVars, checkRun *github.CheckRun) {
	// Add default inputs to those provided by the user
	inputs.workflowInputs["github_repository"] = githubVars.repository
	inputs.workflowInputs["github_sha"] = githubVars.sha
	inputs.workflowInputs["check_id"] = fmt.Sprint(*checkRun.ID)

	rawInputs, err := json.Marshal(inputs.workflowInputs)
	if err != nil {
		githubactions.Fatalf("Error unmarshaling workflow_inputs: %v", err.Error())
	}
	githubactions.Infof("Complete workflow inputs: %v\n", string(rawInputs))
}
