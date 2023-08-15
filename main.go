// @author: wxx
// @since: 2023/7/5
// @desc: // 压缩文件工具
package main

import (
	"bufio"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	ShowTips()
	execute()
	// 如果不点击延时五分钟
	time.Sleep(5 * time.Minute)
}

type ImageFile struct {
	OutPutPath string // 输出目录
	LocalPath  string // 输入目录或者文件路径
	Quality    int    // 质量
	With       int    // 宽度尺寸，像素单位
}

// ShowTips 工具使用步骤
func ShowTips() {
	tips := []string{
		"如果输入文件夹,那么该目录的图片将会被批量压缩\n",
		"例如: 'C:/Users/lzq/Desktop/headImages/ 75 200' ",
		"指桌面 headImages 文件夹，里面的图片质量压缩到75%，宽分辨率为200，高是等比例计算\n",
		"如果是图片路径，那么将会被单独压缩处理\n",
		"例如： 'C:/Users/lzq/Desktop/headImages/1.jpg 75 200' ",
		"指桌面的 headImages 文件夹里面的 1.jpg 图片,质量压缩到75%，宽分辨率为200，高是等比例计算\n",
		"===========>请输入图片或者文件路径，并按回车<===========:\n",
	}
	for k, v := range tips {
		if k == 1 || k == 4 {
			fmt.Println(v)
		} else {
			fmt.Printf(v)
		}
	}
}

// IsImage 判断是否为图片
func IsImage(inputPath string) (path, name, ty string, err error) {
	temp := strings.Split(inputPath, ".")
	if len(temp) <= 1 {
		err = fmt.Errorf("图片命名问题")
		return
	}
	ty = temp[1]
	path = temp[0]
	str := strings.Split(path, "/")
	name = str[len(str)-1]
	switch temp[1] {
	case "jpg":
	case "png":
	case "jpeg":
		return
	default:
		err = fmt.Errorf("%s的图片格式不存在", temp[1])
	}
	return
}

// GetFileList 获取文件
func (i ImageFile) GetFileList(path string) {
	// 创建输出目录
	err := os.MkdirAll(i.OutPutPath, 0777)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	// 遍历目录中的文件
	err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if info == nil {
			return err
		}
		// 是否是目录
		if info.IsDir() {
			return nil
		}
		// 找到一个文件，判断是否为图片
		localPath, _, format, _ := IsImage(path)
		// 随机数
		t := time.Now()
		// 纳秒
		millis := t.Nanosecond()
		// 输出图片path
		outputPath := i.OutPutPath + strconv.FormatInt(int64(millis), 10) + "." + format
		if localPath != "" {
			if !imageCompress(
				func() (io.Reader, error) {
					return os.Open(localPath)
				},
				func() (*os.File, error) {
					return os.Open(localPath)
				},
				outputPath,
				i.Quality,
				i.With,
				format,
			) {
				fmt.Println("生成缩略图失败")
			} else {
				fmt.Println("生成缩略图失败" + outputPath)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("输入的路径信息有误 %v\n", err)
	}
}

// imageCompress 图片压缩
func imageCompress(
	getReadSizeFile func() (io.Reader, error),
	getDecodeFile func() (*os.File, error),
	to string,
	Quality, base int,
	format string,
) bool {
	// 读取文件
	fileOrigin, err := getDecodeFile()
	defer fileOrigin.Close()
	if err != nil {
		fmt.Println("os.Open(file)错误")
		log.Fatal(err)
		return false
	}
	var origin image.Image
	var config image.Config
	var temp io.Reader
	// 读取尺寸
	temp, err = getReadSizeFile()
	if err != nil {
		fmt.Println("os.Open(temp)")
		log.Fatal(err)
		return false
	}
	var typeImage int64
	format = strings.ToLower(format)
	// jpg 格式
	if format == "jpg" || format == "jpeg" {
		fmt.Println("jpg 格式压缩")
		typeImage = 1
		origin, err = jpeg.Decode(fileOrigin)
		if err != nil {
			fmt.Println("jpeg.Decode(file_origin)")
			log.Fatal(err)
			return false
		}
		config, err = jpeg.DecodeConfig(temp)
		fmt.Printf("image config: Width %d;Height %d\n", config.Width, config.Height)
		if err != nil {
			fmt.Println("jpeg.DecodeConfig(temp)")
			log.Fatal(err)
			return false
		}
	} else if format == "png" {
		fmt.Println("png 格式压缩")
		typeImage = 0
		origin, err = png.Decode(fileOrigin)
		if err != nil {
			fmt.Println("png.Decode(file_origin)")
			log.Fatal(err)
			return false
		}
		config, err = png.DecodeConfig(temp)
		if err != nil {
			fmt.Println("png.DecodeConfig(temp)")
			log.Fatal(err)
			return false
		}
	}

	// 基准
	width := uint(base)
	height := uint(base * config.Height / config.Width)
	// 等比例缩放
	canvas := resize.Thumbnail(width, height, origin, resize.Lanczos3)
	fileOut, err := os.Create(to)
	if err != nil {
		fmt.Println("os.Create ", err)
		return false
	}
	if typeImage == 0 {
		err = png.Encode(fileOut, canvas)
		if err != nil {
			fmt.Println("压缩图片失败")
			return false
		}
	} else {
		err = jpeg.Encode(fileOut, canvas, &jpeg.Options{Quality: Quality})
		if err != nil {
			fmt.Println("压缩图片失败")
			return false
		}
	}
	return true
}

func execute() {
	// 获取输入
	reader := bufio.NewReader(os.Stdin)
	data, _, _ := reader.ReadLine()
	// 分割
	strPice := strings.Split(string(data), " ")
	if len(strPice) < 3 {
		fmt.Println("输入有误，参数数量不足，请重新输入或退出程序： ")
		execute()
		return
	}
	inputArgs := ImageFile{
		LocalPath: strPice[0],
	}
	inputArgs.Quality, _ = strconv.Atoi(strPice[1])
	inputArgs.With, _ = strconv.Atoi(strPice[2])

	pathTemp, top, format, _ := IsImage(inputArgs.LocalPath)
	fmt.Printf("local path: %v\n", inputArgs.LocalPath)
	if pathTemp == "" {
		// 目录
		// 如果输入目录，那么是批量
		fmt.Println("开始批量压缩")
		rs := []rune(inputArgs.LocalPath)
		end := len(rs)
		substr := string(rs[end-1 : end])
		if substr == "/" {
			rs = []rune(inputArgs.LocalPath)
			end = len(rs)
			substr = string(rs[0 : end-1])
			endIndex := strings.LastIndex(substr, "/")
			inputArgs.OutPutPath = string(rs[0:endIndex]) + "/LghImageCompress/"
		} else {
			endIndex := strings.LastIndex(inputArgs.LocalPath, "/")
			inputArgs.OutPutPath = string(rs[0:endIndex]) + "/LghImageCompress/"
		}
		inputArgs.GetFileList(inputArgs.LocalPath)
	} else {
		// 单个
		// 如果输入文件，要么是单个，允许自定义路径
		fmt.Println("开始单张压缩...")
		inputArgs.OutPutPath = top + "_compress." + format
		if !imageCompress(func() (io.Reader, error) {
			return os.Open(inputArgs.LocalPath)
		}, func() (*os.File, error) {
			return os.Open(inputArgs.LocalPath)
		},
			inputArgs.OutPutPath,
			inputArgs.Quality,
			inputArgs.With, format) {
			fmt.Println("生成缩略图失败")
		} else {
			fmt.Println("生成缩略图成功" + inputArgs.OutPutPath)
			finish()
		}

	}
}

func finish() {
	fmt.Printf("继续输入进行压缩或者退出程序：")
	execute()
}
