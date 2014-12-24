package main

import (
	"fmt"
	"os/exec"
	"flag"
	"bytes"
	"encoding/json"
	"log"
	"io/ioutil"
	"strings"
	"strconv"
	"github.com/mattn/go-jsonpointer"
	"github.com/cmiceli/password-generator-go"
)

type awsPolicy struct {
	GroupName			string		`json:"GroupName"`
	PolicyName		string		`json:"PolicyName"`
	PolicyPath		string		`json:"PolicyPath"`
}

type awsParam struct {
	policy				awsPolicy
	userName			[]string
	awsAccount		string
}

func main() {
	var userNames string
	var param awsParam
	var policyJson string
	flag.StringVar(&userNames, "user-name", "", "user-name1,user-name2")
	flag.StringVar(&param.awsAccount, "account", "", ".aws/config file account")
	flag.StringVar(&policyJson	, "policy-json", "", "policy setting json file path")
	flag.Parse()

	//get json parameter
	param.policy = getPolicyJson(policyJson)
	param.userName = strings.Split(userNames, ",")

	awsAccountGenerate(param)
}

func awsAccountGenerate(param awsParam) {
	for _, v := range param.userName {
		cmdCreateGroupAndPolicy(param.policy.GroupName, param.policy.PolicyName, param.policy.PolicyPath, param.awsAccount)

		fmt.Println("URL: ")
		fmt.Println(cmdAccountAlias(param.awsAccount))

		fmt.Println("userName: ")
		fmt.Println(cmdCreateUser(param.awsAccount, v))
		cmdAddUserToGroup(param.policy.GroupName, v, param.awsAccount)

		fmt.Println("password: ")
		fmt.Println(cmdSetUserPassword(param.awsAccount, v))

		keyId, secretKey := cmdAccessKey(param.awsAccount, v)
		fmt.Println("accessKeyId: ")
		fmt.Println(keyId)
		fmt.Println("secretAccessKey: ")
		fmt.Println(secretKey)

		fmt.Println()
	}

}

func getPolicyJson(policyJsonPath string) (policy awsPolicy) {
	policyJson, err := ioutil.ReadFile(policyJsonPath)

	if err != nil {
		log.Fatalln("policy json read error:",err)
	}

	err = json.Unmarshal(policyJson, &policy)
	if err != nil {
		log.Fatalln("json unmarshal error:",err)
	}

	return
}

func cmdRun(cmd *exec.Cmd) (resultJson interface{}, err error) {
	var commandResponse bytes.Buffer
	cmd.Stdout = &commandResponse
	err = cmd.Run()
	if err != nil {
		fmt.Println(cmd.Args)
		log.Fatalln("cmd.Run:",err)
	}

	if commandResponse.Len() == 0 {
		return
	}

	err = json.Unmarshal([]byte(commandResponse.Bytes()), &resultJson)
	if err != nil {
		fmt.Println(cmd.Args)
		log.Fatalln("json.Unmarshal:",err)
	}

	return
}

func cmdCreateGroupAndPolicy(group string, policy string, path string,  account string) {
	execCommand := exec.Command("aws", "iam", "list-groups", "--profile", account)
	groupsJson, _ := cmdRun(execCommand)

	for i := 0; ; i++ {
		isValid := jsonpointer.Has(groupsJson, "/Groups/" + strconv.Itoa(i) + "/GroupName")
		if isValid == false {
			break
		}

		getGroup, _ := jsonpointer.Get(groupsJson, "/Groups/" + strconv.Itoa(i) + "/GroupName")
		if getGroup == group {
			//fmt.Println("group [" + group + "] is already.")
			return
		}
	}

	execCommand = exec.Command("aws", "iam", "create-group", "--group-name", group, "--profile", account)
	_, err := cmdRun(execCommand)
	if err != nil {
		log.Fatalln("create-group:",err)
	}

	execCommand = exec.Command("aws", "iam", "put-group-policy", "--group-name", group, "--policy-name", policy, "--policy-document", path, "--profile", account)
	_, err = cmdRun(execCommand)
	if err != nil {
		log.Fatalln("put-group-policy:",err)
	}
}

func cmdAddUserToGroup(group string, user string, account string) {
	execCommand := exec.Command("aws", "iam", "add-user-to-group", "--group-name", group, "--user-name", user, "--profile", account)
	_, err := cmdRun(execCommand)

	if err != nil {
		log.Fatalln("add-user-to-group:",err)
	}
}

func cmdAccountAlias(account string) string {
	execCommand :=	exec.Command("aws", "iam", "list-account-aliases", "--profile", account)
	result, _ := cmdRun(execCommand)
	accountAlias, _ := jsonpointer.Get(result, "/AccountAliases")

	if nil == accountAlias {
		execCommand :=	exec.Command("aws", "iam", "create-account-alias", "--account-alias", account, "--profile", account)
		result, _ = cmdRun(execCommand)
		accountAlias, _ = jsonpointer.Get(result, "/AccountAliases")
	} else {
		//fmt.Println("create account alias is already.")
	}

	a := fmt.Sprint(accountAlias)
	return "https://" + a[1:len(a) - 1] + ".signin.aws.amazon.com/console"
}

func cmdCreateUser(account string, user string) interface{} {
	execCommand :=	exec.Command("aws", "iam", "create-user", "--user-name", user, "--profile", account)
	result, _ := cmdRun(execCommand)
	userName, _ := jsonpointer.Get(result, "/User/UserName")
	return userName
}

func cmdSetUserPassword(account string, user string) string {
	password := pwordgen.NewPassword(20)
	execCommand :=	exec.Command("aws", "iam", "create-login-profile", "--user-name", user, "--password", password, "--profile", account)
	_, err := cmdRun(execCommand)
	if err != nil {
		log.Fatalln("cmdSetUserPassword",err)
	}
	return password
}

func cmdAccessKey(account string, user string) (interface{}, interface{}) {
	execCommand :=	exec.Command("aws", "iam", "create-access-key", "--user-name", user, "--profile", account)
	result, _ := cmdRun(execCommand)
	keyId, _ := jsonpointer.Get(result, "/AccessKey/AccessKeyId")
	secretKey, _ := jsonpointer.Get(result, "/AccessKey/SecretAccessKey")
	return keyId, secretKey
}
