/*
Copyright © 2020 Alessandro Segala (@ItalyPaleAle)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/ItalyPaleAle/prvt/cmd"
	"github.com/stretchr/testify/assert"
)

// RunServer runs the sequence of tests for the server; must be run run after the CLI tests
func (s *funcTestSuite) RunServer(t *testing.T) {
	var close func()
	// Test on existing repo
	s.promptPwd.SetPasswords("hello world")
	close = s.startServer(t, "--store", "local:"+s.dirs[0])
	t.Run("info", s.serverInfo)
	t.Run("add files", s.serverAdd)
	close()
}

func (s *funcTestSuite) startServer(t *testing.T, args ...string) func() {
	// Start the server by invoking the command
	serveCmd := cmd.NewServeCmd()
	serveCmd.SetOut(ioutil.Discard)
	serveCmd.SetErr(ioutil.Discard)
	serveCmd.SetArgs(args)

	// Server address
	address, err := serveCmd.Flags().GetString("address")
	if err != nil {
		t.Fatal(err)
		return nil
	}
	port, err := serveCmd.Flags().GetString("port")
	if err != nil {
		t.Fatal(err)
		return nil
	}
	s.serverAddr = fmt.Sprintf("http://%s:%s", address, port)

	// Set a context that can be canceled, used to stop the server
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := serveCmd.ExecuteContext(ctx)
		if err != nil {
			log.Fatal("cannot start server\nargs:", args, "\nerror:", err)
			return
		}
	}()

	// Wait a couple of seconds to ensure the server has started
	time.Sleep(2 * time.Second)

	// The caller can stop the server
	return cancel
}

func (s *funcTestSuite) serverInfo(t *testing.T) {
	sendRequest := func() (data map[string]string, err error) {
		var (
			res  *http.Response
			read []byte
		)

		// Send the request, then read the response and parse the JSON response into a map
		res, err = s.client.Get(s.serverAddr + "/api/info")
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		read, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		data = make(map[string]string)
		err = json.Unmarshal(read, &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	// Check the response
	data, err := sendRequest()
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, data, 2)
	assert.Equal(t, "prvt", data["name"])
	assert.True(t, len(data["info"]) > 0)

	// Set buildinfo then check again
	reset := setBuildInfo()
	data, err = sendRequest()
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, data, 6)
	assert.Equal(t, "prvt", data["name"])
	assert.True(t, len(data["version"]) > 0)
	assert.True(t, len(data["buildId"]) > 0)
	assert.True(t, len(data["buildTime"]) > 0)
	assert.True(t, len(data["commitHash"]) > 0)
	assert.True(t, len(data["runtime"]) > 0)
	reset()
}

func (s *funcTestSuite) serverAdd(t *testing.T) {
	// TODO
}
