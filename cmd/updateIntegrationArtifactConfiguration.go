package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/command"
	cpi "github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type updateIntegrationArtifactConfigurationUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The updateIntegrationArtifactConfigurationUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type updateIntegrationArtifactConfigurationUtilsBundle struct {
	*command.Command
	// Embed more structs as necessary to implement methods or interfaces you add to updateIntegrationArtifactConfigurationUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// updateIntegrationArtifactConfigurationUtilsBundle and forward to the implementation of the dependency.
}

func newUpdateIntegrationArtifactConfigurationUtils() updateIntegrationArtifactConfigurationUtils {
	utils := updateIntegrationArtifactConfigurationUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func updateIntegrationArtifactConfiguration(config updateIntegrationArtifactConfigurationOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}
	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runUpdateIntegrationArtifactConfiguration(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runUpdateIntegrationArtifactConfiguration(config *updateIntegrationArtifactConfigurationOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) error {
	clientOptions := piperhttp.ClientOptions{}
	httpClient.SetOptions(clientOptions)

	configUpdateURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')/$links/Configurations('%s')", config.Host, config.IntegrationFlowID, config.IntegrationFlowVersion, config.ParameterKey)
	tokenParameters := cpi.TokenParameters{TokenURL: config.OAuthTokenProviderURL, Username: config.Username, Password: config.Password, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "PUT"
	header := make(http.Header)
	header.Add("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	jsonObj := gabs.New()
	jsonObj.Set(config.ParameterValue, "ParameterValue")
	jsonBody, httpErr := json.Marshal(jsonObj)

	if httpErr != nil {
		return errors.Wrap(httpErr, "input json body is invalid")
	}
	configUpdateResp, httpErr := httpClient.SendRequest(httpMethod, configUpdateURL, bytes.NewBuffer(jsonBody), header, nil)
	if httpErr != nil {
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", httpMethod, configUpdateURL)
	}

	if configUpdateResp != nil && configUpdateResp.Body != nil {
		defer configUpdateResp.Body.Close()
	}

	if configUpdateResp == nil {
		return errors.Errorf("did not retrieve a HTTP response")
	}

	if configUpdateResp.StatusCode == 202 {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("successfully updated the integration flow configuration parameter")
		return nil
	}

	log.Entry().Errorf("a HTTP error occurred! Response Status Code: %v", configUpdateResp.StatusCode)
	return errors.Errorf("Failed to update the integration flow configuration parameter, Response Status code: %v", configUpdateResp.StatusCode)
}
