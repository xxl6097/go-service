package ukey

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/utils"
	"io"
	"math"
	"os"
	"path/filepath"
)

func SignFileBySelfKey(cfg any, inFilePath string) (string, error) {
	//1、获取用户的配置信息；
	//2、加密用户信息，构建二进制常量；
	//3、从原始二进制文件中查询特征码，替换为二进制常量；
	//4、替换后，将二进制文件复制到安装指定的目录
	newBufferBytes, err := GenConfig(cfg, false)
	if err != nil {
		return "", fmt.Errorf("构建签名信息错误: %v", err)
	}
	outFilePath := filepath.Join(glog.GetCrossPlatformDataDir("sign", utils.GetID()), filepath.Base(inFilePath))
	//安装程序，需要对程序进行签名，那么需要传入两个参数：
	//1、最原始的key；
	//2、需写入的data
	buf := GetBuffer()
	glog.Printf("buffer大小 %d\n", len(buf))
	err = GenerateBin(inFilePath, outFilePath, buf, newBufferBytes)
	if err != nil {
		return "", fmt.Errorf("签名错误: %v", err)
	}
	return outFilePath, nil
}

func SignFileByOldFileKey(oldFilePath, newFilePath string) (string, error) {
	glog.Debugf("\n旧文件：%s\n新文件：%s\n", oldFilePath, newFilePath)
	//1、读取老文件特征数据；
	//2、下载新文件
	//3、替换新文件特征数据
	//4、数据写到安装目录地址（oldBinPath）
	cfgBufferBytes := GetCfgBufferFromFile(oldFilePath)
	if cfgBufferBytes == nil {
		err := fmt.Errorf("读取原文件配置信息失败 %s", oldFilePath)
		glog.Error(err)
		return "", err
	}
	outFilePath := filepath.Join(glog.GetCrossPlatformDataDir("sign", utils.GetID()), filepath.Base(newFilePath))
	glog.Debug("获取配置数据成功", len(cfgBufferBytes))
	oldBuffer := GetBuffer()
	err := GenerateBin(newFilePath, outFilePath, oldBuffer, cfgBufferBytes)
	if err != nil {
		glog.Error("签名错误：", err)
		return "", err
	}
	return outFilePath, nil
}

func GenerateBin(scrFilePath, dstFilePath string, oldBytes, newBytes []byte) error {
	// 打开原文件
	srcFile, err := os.Open(scrFilePath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %v[%s]", err, scrFilePath)
	}
	defer srcFile.Close()

	var srcFileSize int64
	if stat, err := srcFile.Stat(); err == nil {
		srcFileSize = stat.Size()
		sizeB := float64(stat.Size()) / 1024 / 1024
		glog.Printf("%s[大小：%.2fMB]%s\n", stat.Name(), sizeB, dstFilePath)
	}

	tmpFile, err := os.Create(dstFilePath)
	if err != nil {
		return fmt.Errorf("无法创建临时文件: %v[%s]", err, dstFilePath)
	}
	defer tmpFile.Close()

	reader := bufio.NewReader(srcFile)
	prevBuffer := make([]byte, 0)
	isReplace := false
	var indexSize int64
	newFileSize := int64(0)
	tempProgress := -1
	for {
		thisBuffer := make([]byte, Divide(len(oldBytes), 1024))
		n, err2 := reader.Read(thisBuffer)
		if err2 != nil && err2 != io.EOF {
			return fmt.Errorf("读取文件时出错: %v[%s]", err2, scrFilePath)
		}
		indexSize += int64(n)
		thisBuffer = thisBuffer[:n]
		tempBuffer := append(prevBuffer, thisBuffer...)
		index := bytes.Index(tempBuffer, oldBytes)
		if index > -1 {
			//glog.Printf("找到位置[%d]了，签名...\n", index)
			glog.Printf("程序签名成功[%d]\n", index)
			isReplace = true
			tempBuffer = bytes.Replace(tempBuffer, oldBytes, newBytes, -1)
		}
		// 写入前一次的
		writeSize, err1 := tmpFile.Write(tempBuffer[:len(prevBuffer)])
		if err1 != nil {
			return fmt.Errorf("1写入临时文件时出错: %v[%s]", err1, dstFilePath)
		}

		newFileSize += int64(writeSize)
		progress := int(float64(indexSize) / float64(srcFileSize) * 100)
		if progress >= tempProgress {
			//glog.Printf("程序签名:%v%s\n", progress, "%")
			tempProgress = progress
			tempProgress += 5
		}

		//前一次的+本次的转给 prev
		prevBuffer = tempBuffer[len(prevBuffer):]
		//if err != nil {
		//	break
		//}
		if n == 0 || err2 != nil {
			break // 文件读取完毕
		}
	}
	if len(prevBuffer) > 0 {
		writeSize, err1 := tmpFile.Write(prevBuffer)
		if err1 != nil {
			return fmt.Errorf("2写入临时文件时出错: %v[%s]", err1, dstFilePath)
		}
		newFileSize += int64(writeSize)
		prevBuffer = nil
	}
	glog.Printf("原始文件大小：%d  %s\n", indexSize, scrFilePath)
	glog.Printf("目标文件大小：%d  %s\n", indexSize, dstFilePath)
	// 给文件赋予执行权限（0755）
	errMsg := os.Chmod(dstFilePath, 0755)
	if errMsg != nil {
		return fmt.Errorf("赋予文件执行权限时出错: %v\n", errMsg)
	}
	if !isReplace {
		glog.Printf("oldBytes[%d]--->%v\n", len(oldBytes), oldBytes)
		glog.Printf("newBytes[%d]--->%v\n", len(newBytes), newBytes)
		return errors.New("位置没找到，数据未替换😭")
	}
	err1 := srcFile.Close()
	if err1 != nil {
		glog.Error("srcFile.Close", err1)
	}
	err1 = tmpFile.Close()
	if err1 != nil {
		glog.Error("tmpFile.Close", err1)
	}

	return nil
}

// DivideAndCeil 函数用于进行除法并向上取整
func DivideAndCeil(a, b int) int {
	// 将整数转换为 float64 类型进行除法运算
	result := float64(a) / float64(b)
	// 使用 math.Ceil 函数进行向上取整
	result = math.Ceil(result)
	// 将结果转换回整数类型
	return int(result)
}

func Divide(a, b int) int {
	return DivideAndCeil(a, b) * b
}
