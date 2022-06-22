/*
Copyright 2018 The pdfcpu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package main provides the command line for interacting with pdfcpu.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/cli"
)

var (
	fileStats, mode, selectedPages  string
	upw, opw, key, perm, unit, conf string
	verbose, veryVerbose            bool
	links, quiet, sorted            bool
	needStackTrace                  = true
	cmdMap                          commandMap
)

// Set by Goreleaser.
var (
	commit = "?"
	date   = "?"
	outDir = "out_temp"
)

func init() {
	initFlags()
	initCommandMap()
}

func main() {
	conf, err := ensureDefaultConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(0)
	}
	conf.OwnerPW = opw
	conf.UserPW = upw
	var inputReader *bufio.Reader = bufio.NewReader(os.Stdin)

	if len(os.Args) > 2 {
		// 合并

		filesIn := []string{}
		// 从第二个参数开始，是要合并的页数
		for i := 1; i < len(os.Args); i++ {
			filesIn = append(filesIn, os.Args[i])
		}

		// 打印出来，看看对不对
		fmt.Println(filesIn)

		outFile := "out.pdf"
		// 合并
		cmd := cli.MergeCreateCommand(filesIn, outFile, conf)
		cli.Process(cmd)
		fmt.Println("结束")
		inputReader.ReadString('\n')
		os.Exit(0)
	}

	// 切分
	// 入参，要切割的文件
	inFile := os.Args[1]
	// 校验文件是否存在
	if conf.CheckFileNameExt {
		ensurePDFExtension(inFile)
	}
	// 输出目录
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, os.ModePerm)

	f, err := os.Open(inFile)
	if err != nil {
		log.Fatalln(err.Error())
	}

	func() {
		defer func() {
			if err != nil {
				f.Close()
				return
			}
			err = f.Close()
		}()

		api.Split(f, outDir, "out", 1, conf)
	}()

	fmt.Printf("输入需要的页(空格分隔, -表示连续)：")
	input, err := inputReader.ReadString('\n')
	if err == nil {
		fmt.Printf("The input was: %s\n", input)
	}
	println(input)
	items := strings.Split(input, " ")

	// 合并

	filesIn := []string{}
	// 从第二个参数开始，是要合并的页数
	for _, item := range items {
		item = strings.Trim(item, " \r\n\t")
		if strings.Contains(item, "-") {
			ii := strings.Split(item, "-")
			s, _ := strconv.Atoi(ii[0])
			e, _ := strconv.Atoi(ii[1])
			for i := s; i <= e; i++ {
				filesIn = append(filesIn, outDir+"\\out_"+strconv.Itoa(i)+".pdf")
			}
		} else {
			filesIn = append(filesIn, outDir+"\\out_"+item+".pdf")
		}
	}

	// 打印出来，看看对不对
	fmt.Println(filesIn)

	outFile := "out.pdf"
	// 合并
	cmd := cli.MergeCreateCommand(filesIn, outFile, conf)
	out, err := cli.Process(cmd)
	if err != nil {
		if needStackTrace {
			fmt.Fprintf(os.Stderr, "Fatal: %+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		os.Exit(1)
	}

	if out != nil && !quiet {
		for _, s := range out {
			fmt.Fprintln(os.Stdout, s)
		}
	}

	os.RemoveAll(outDir)

	fmt.Println("结束")
	inputReader.ReadString('\n')
}
