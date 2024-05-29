package notification

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"rmk/config"
)

type SlackTmp interface {
	TmpProjectUpdateMsg() string
	TmpReleaseUpdateMsg() string
	TmpReleaseUpdateSuccessMsg() string
	TmpReleaseUpdateFailedMsg(err error) string
	TmpUpdateMsgDetails() string
}

type TmpUpdate struct {
	ChangesList []string
	PathToFile  string
	*cli.Context
	*config.Config
}

func (t *TmpUpdate) TmpProjectUpdateMsg() string {
	return fmt.Sprintf("*Dependency:* _%s_\n"+
		"*New version:* `%s`\n"+
		"\t*Affected tenant:* %s\n"+
		"\t*Affected environment:* %s\n"+
		"\t*Affected file:* %s\n",
		strings.Join(t.ChangesList, " "),
		t.Context.String("version"),
		t.Tenant,
		t.Environment,
		t.PathToFile,
	) + t.TmpUpdateMsgDetails()
}

func (t *TmpUpdate) TmpReleaseUpdateSuccessMsg() string {
	return fmt.Sprintf("*Success deployed releases:* _%s_\n"+
		"*For cluster:* %s\n",
		strings.Join(t.ChangesList, " "),
		t.RootDomain,
	) + t.TmpUpdateMsgDetails()
}

func (t *TmpUpdate) TmpReleaseUpdateFailedMsg(err error) string {
	return fmt.Sprintf("*Fail deployed releases:* _%s_\n"+
		"*For cluster:* %s\n"+
		"*Error:* %s\n",
		strings.Join(t.ChangesList, " "),
		t.RootDomain,
		err,
	) + t.TmpUpdateMsgDetails()
}

func (t *TmpUpdate) TmpReleaseUpdateMsg() string {
	scope := strings.Split(t.PathToFile, string(filepath.Separator))

	return fmt.Sprintf("*Releases:* _%s_\n"+
		"*New version:* `%s`\n"+
		"\t*Affected tenant:* %s\n"+
		"\t*Affected environment:* %s\n"+
		"\t*Affected scope:* %s\n",
		strings.Join(t.ChangesList, " "),
		t.Context.String("tag"),
		t.Tenant,
		t.Environment,
		scope[len(scope)-3],
	) + t.TmpUpdateMsgDetails()
}

func (t *TmpUpdate) TmpUpdateMsgDetails() string {
	if len(t.SlackMsgDetails) > 0 {
		return fmt.Sprintf("*Details:*\n\t- %s", strings.Join(t.SlackMsgDetails, "\n\t- "))
	}

	return ""
}
