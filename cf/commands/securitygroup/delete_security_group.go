package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/securitygroups"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type DeleteSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo securitygroups.SecurityGroupRepo
	configRepo        coreconfig.Reader
}

func init() {
	commandregistry.Register(&DeleteSecurityGroup{})
}

func (cmd *DeleteSecurityGroup) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-security-group",
		Description: T("Deletes a security group"),
		Usage: []string{
			T("CF_NAME delete-security-group SECURITY_GROUP [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-security-group"))
	}

	reqs := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return reqs
}

func (cmd *DeleteSecurityGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	return cmd
}

func (cmd *DeleteSecurityGroup) Execute(context flags.FlagContext) error {
	name := context.Args()[0]
	cmd.ui.Say(T("Deleting security group {{.security_group}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	if !context.Bool("f") {
		response := cmd.ui.ConfirmDelete(T("security group"), name)
		if !response {
			return nil
		}
	}

	group, err := cmd.securityGroupRepo.Read(name)
	switch err.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Security group {{.security_group}} does not exist", map[string]interface{}{"security_group": name}))
		return nil
	default:
		return err
	}

	err = cmd.securityGroupRepo.Delete(group.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
