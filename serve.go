package main

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func toPtr(s string) *uint16 {
	if len(s) == 0 {
		return nil
	}
	return syscall.StringToUTF16Ptr(s)
}

// StartProcess ...
func StartProcess(appPath, cmdLine, workDir string) error {
	fmt.Println("start process:", appPath, cmdLine, workDir)
	var (
		hChildStd_IN_Rd  windows.Handle
		hChildStd_IN_Wr  windows.Handle
		hChildStd_OUT_Rd windows.Handle
		hChildStd_OUT_Wr windows.Handle

		err error
	)

	saAttr := &windows.SecurityAttributes{
		InheritHandle: 1,
	}

	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return fmt.Errorf("CreateJobObject err: %v", err)
	}
	defer windows.CloseHandle(job)

	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}

	_, err = windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)))
	if err != nil {
		return fmt.Errorf("SetInformationJobObject err:%v", err)
	}

	// var HandleFlagInherit uint32 = 0x00000001
	var NotHandleFlagInherit uint32 = 0x00000000

	// Create a pipe for the child process's STDOUT
	err = windows.CreatePipe(&hChildStd_OUT_Rd, &hChildStd_OUT_Wr, saAttr, 0)
	if err != nil {
		return fmt.Errorf("CreatePipe err:%v", err)
	}

	// Ensure the read handle to the pipe for STDOUT is not inherited
	err = windows.SetHandleInformation(hChildStd_OUT_Rd, NotHandleFlagInherit, 0)
	if err != nil {
		return fmt.Errorf("SetHandleInformation err:%v", err)
	}

	// Create a pipe for the child process's STDIN
	err = windows.CreatePipe(&hChildStd_IN_Rd, &hChildStd_IN_Wr, saAttr, 0)
	if err != nil {
		return fmt.Errorf("CreatePie err:%v", err)
	}

	// Ensure the write handle to the pipe for STDIN is not inherited
	err = windows.SetHandleInformation(hChildStd_IN_Wr, NotHandleFlagInherit, 0)
	if err != nil {
		return fmt.Errorf("SetHandleInformation err:%v", err)
	}

	// if len(cmdLine) > 0 {
	// 	commandLine = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(cmdLine)))
	// }
	// if len(workDir) > 0 {
	// 	workingDir = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(workDir)))
	// }

	processInfo := windows.ProcessInformation{}
	startupInfo := windows.StartupInfo{
		StdErr:    hChildStd_OUT_Wr,
		StdOutput: hChildStd_OUT_Wr,
		StdInput:  hChildStd_IN_Rd,
		Flags:     windows.STARTF_USESTDHANDLES,
	}
	creationFlags := uint32(windows.CREATE_NO_WINDOW | windows.CREATE_UNICODE_ENVIRONMENT)

	// returnCode, _, err := procCreateProcessAsUser.Call(uintptr(userToken), uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(appPath))), commandLine, 0, 0, 1,
	// uintptr(creationFlags), uintptr(envInfo), workingDir, uintptr(unsafe.Pointer(&startupInfo)), uintptr(unsafe.Pointer(&processInfo)))
	// returnCode, _, err := procCreateProcess.Call(
	err = windows.CreateProcess(
		toPtr(appPath),
		toPtr(cmdLine),
		nil,
		nil,
		true,
		creationFlags,
		nil,
		toPtr(workDir),
		&startupInfo,
		&processInfo)
	// if returnCode == 0 {
	if err != nil {
		return fmt.Errorf("Unable to create process: %s", err)
	}

	err = windows.AssignProcessToJobObject(job, processInfo.Process)
	if err != nil {
		return fmt.Errorf("AssignProcessToJobObject err: %v", err)
	}

	err = windows.CloseHandle(hChildStd_OUT_Wr)
	fmt.Println("close err:", err)
	err = windows.CloseHandle(hChildStd_IN_Rd)
	fmt.Println("close err:", err)

	overlappedRx := windows.Overlapped{}
	overlappedRx.HEvent, err = windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		fmt.Println("err:", err)
	}

	buf := make([]byte, 1024)
	var done uint32

	for {
		err = windows.ReadFile(hChildStd_OUT_Rd, buf, &done, &overlappedRx)
		if err != nil {
			fmt.Printf("Unable to ReadFile: %v\n", err)
			break
		}
		fmt.Printf("%s", string(buf))
		if _, err := windows.WaitForSingleObject(overlappedRx.HEvent, windows.INFINITE); err != nil {
			fmt.Println(err)
		}

		overlappedRx.Offset += done
	}
	return err
}

// func main() {
// 	app := "C:\\ProgramData\\chocolatey\\bin\\pulseaudio.exe"
// 	// app := "C:\\Windows\\notepad.exe"
// 	cmd := ""

// 	err := StartProcess(app, cmd, "")
// 	if err != nil {
// 		fmt.Println("err:", err)
// 		err = syscall.GetLastError()
// 		fmt.Println("GetLastError:", err)
// 	}
// }

// start pulseaudio.exe
func serve() {

	app := "C:\\ProgramData\\chocolatey\\bin\\pulseaudio.exe"
	cmd := ""

	err := StartProcess(app, cmd, "")
	if err != nil {
		fmt.Println("err:", err)
		err = syscall.GetLastError()
		fmt.Println("GetLastError:", err)
		return
	}

	return
}

func ShowOkMessageBox(title, text string) {
	log.Print(text)
	_, _ = windows.MessageBox(0, toPtr(text), toPtr(title), windows.MB_ICONINFORMATION)
}
