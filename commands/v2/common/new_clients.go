package common

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/commands"
)

func NewClients(config commands.Config, ui TerminalDisplay) (*ccv2.Client, *uaa.Client, error) {
	if config.Target() == "" {
		return nil, nil, NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	ccClient := NewCloudControllerClient(config.BinaryName())
	_, err := ccClient.TargetCF(ccv2.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return nil, nil, err
	}

	uaaClient := uaa.NewClient(uaa.Config{
		DialTimeout:       config.DialTimeout(),
		SkipSSLValidation: config.SkipSSLValidation(),
		Store:             config,
		URL:               ccClient.TokenEndpoint(),
	})

	verbose, location := config.Verbose()
	if verbose {
		logger := wrapper.NewRequestLogger(NewRequestLoggerTerminalDisplay(ui))
		ccClient.WrapConnection(logger)
	}
	if location != nil {
		logger := wrapper.NewRequestLogger(NewRequestLoggerFileWriter(ui, location))
		ccClient.WrapConnection(logger)
	}

	ccClient.WrapConnection(wrapper.NewUAAAuthentication(uaaClient))
	ccClient.WrapConnection(wrapper.NewRetryRequest(2))
	return ccClient, uaaClient, err
}

func NewCloudControllerClient(binaryName string) *ccv2.Client {
	return ccv2.NewClient(binaryName, cf.Version)
}
