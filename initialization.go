package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v37/github"
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
	workflowInputs     map[string]interface{}
}

func getInputs() (inputs, error) {

	appIDString, ok := os.LookupEnv("INPUT_APP_ID")
	if !ok {
		return inputs{}, errors.New("input 'app_id' not set")
	}
	appID, err := strconv.ParseInt(appIDString, 10, 64)
	if err != nil {
		return inputs{}, errors.New("input app_id must be an integer")
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
	if err != nil {
		return inputs{}, errors.New("input 'target_repository' not set")
	}

	targetOwner, ok := os.LookupEnv("INPUT_TARGET_OWNER")
	if err != nil {
		return inputs{}, errors.New("input 'target_owner' not set")
	}

	targetRef, ok := os.LookupEnv("INPUT_TARGET_REF")
	if err != nil {
		return inputs{}, errors.New("input 'target_ref' not set")
	}

	workflowFilename, ok := os.LookupEnv("WORKFLOW_FILENAME")
	if !ok {
		return inputs{}, errors.New("input 'workflow_filename' not set")
	}

	wfcString := os.Getenv("INPUT_WAIT_FOR_CHECK")
	waitForCheck, err := strconv.ParseBool(wfcString)
	if err != nil {
		return inputs{}, fmt.Errorf("input 'wait_for_check' is not a boolean: %w", err)
	}

	waitTimeoutSecondsString, ok := os.LookupEnv("INPUT_WAIT_TIMEOUT_SECONDS")
	waitTimeoutSeconds, err := strconv.ParseInt(waitTimeoutSecondsString, 10, 64)
	if err != nil {
		return inputs{}, errors.New("input wait_timeout_seconds must be an integer")
	}

	workflowInputs := map[string]interface{}{}
	workflowInputsString, ok := os.LookupEnv("INPUT_WORKFLOW_INPUTS")
	if !ok {
		return inputs{}, errors.New("input 'client_payload' not set")
	}
	err = json.Unmarshal([]byte(workflowInputsString), &workflowInputs)
	if err != nil {
		return inputs{}, fmt.Errorf("input 'workflowInputs' is not json: %w", err)
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
	}, nil
}

func newGithubClient(tr http.RoundTripper, appID int64, privateKey *rsa.PrivateKey) (*github.Client, error) {
	// https://github.com/google/go-github#authentication
	// First, create an AppsTransport for initial auth
	itr := ghinstallation.NewAppsTransportFromPrivateKey(tr, appID, privateKey)
	baseURL, ok := os.LookupEnv("GITHUB_API_URL")
	if !ok {
		return nil, errors.New("env var 'GITHUB_API_URL' is not set")
	}
	itr.BaseURL = baseURL

	// use appTransport to generate a client
	client := github.NewClient(&http.Client{Transport: itr})

	// Get the list of installations
	opt := &github.ListOptions{
		PerPage: 100,
	}
	var allInstallations []*github.Installation
	for {
		installations, resp, err := client.Apps.ListInstallations(context.Background(), opt)
		if err != nil {
			return nil, fmt.Errorf("error getting installations: %w", err)
		}
		allInstallations = append(allInstallations, installations...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// search for the specific installation we care about
	// spew.Dump(allInstallations)

	// construct client with the installation
	ntr := ghinstallation.NewFromAppsTransport(itr, *allInstallations[0].ID)
	return github.NewClient(&http.Client{Transport: ntr}), nil
}
