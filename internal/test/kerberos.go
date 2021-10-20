package test

import (
	"embed"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

var krbLock = &sync.Mutex{}

//go:embed krb
var krbBuildRoot embed.FS

func krbBuildRootFiles() map[string][]byte {
	files := krbBuildRootDirFiles("krb")
	result := map[string][]byte{}
	for file, data := range files {
		result[strings.TrimPrefix(file, "krb/")] = data
	}
	return result
}

func krbBuildRootDirFiles(dir string) map[string][]byte {
	result := map[string][]byte{}
	fsEntries, err := krbBuildRoot.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, fsEntry := range fsEntries {
		fullPath := dir + "/" + fsEntry.Name()
		if fsEntry.IsDir() {
			subDirFiles := krbBuildRootDirFiles(fullPath)
			for fileName, fileContent := range subDirFiles {
				result[fileName] = fileContent
			}
		} else {
			data, err := krbBuildRoot.ReadFile(fullPath)
			if err != nil {
				panic(err)
			}
			result[fullPath] = data
		}
	}
	return result
}

func Kerberos(t *testing.T) KerberosHelper {
	krbLock.Lock()
	t.Cleanup(
		func() {
			krbLock.Unlock()
		})
	adminUsername := "admin"
	adminPassword := "testing"

	files := krbBuildRootFiles()
	cnt := containerFromBuild(
		t,
		"krb",
		files,
		nil,
		[]string{
			fmt.Sprintf("KERBEROS_USERNAME=%s", adminUsername),
			fmt.Sprintf("KERBEROS_PASSWORD=%s", adminPassword),
		},
		map[string]string{
			"88/tcp": "88",
			"88/udp": "88",
			"464/tcp": "464",
			"464/udp": "464",
			"750/tcp": "750",
			"750/udp": "750",
		},
	)

	helper := &kerberosHelper{
		t: t,
		cnt: cnt,
		adminUsername: adminUsername,
		adminPassword: adminPassword,
	}

	helper.wait()

	return helper
}

type KerberosHelper interface {
	// Realm returns the Kerberos realm to authenticate with.
	Realm() string
	// AdminUsername returns a Kerberos username which has admin privileges.
	AdminUsername() string
	// AdminPassword returns the password for the admin username from AdminUsername.
	AdminPassword() string
}

type kerberosHelper struct {
	adminUsername string
	adminPassword string
	cnt           container
	t             *testing.T
}

func (k *kerberosHelper) Realm() string {
	return "TESTING.CONTAINERSSH.IO"
}

func (k *kerberosHelper) AdminUsername() string {
	return k.adminUsername
}

func (k *kerberosHelper) AdminPassword() string {
	return k.adminPassword
}

func (k *kerberosHelper) wait() {
	k.t.Log("Waiting for the KDC to come up...")
	tries := 0
	sleepTime := 5
	for {
		if tries > 30 {
			k.t.Fatalf("The KDC failed to come up in %d seconds", sleepTime * 30)
		}
		sock, err := net.Dial("tcp", "127.0.0.1:88")
		time.Sleep(time.Duration(sleepTime) * time.Second)
		if err != nil {
			tries++
		} else {
			_ = sock.Close()

			return
		}
	}
}
