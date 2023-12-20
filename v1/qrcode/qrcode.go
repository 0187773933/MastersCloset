package qrcode

import (
	"fmt"
	"os"
	"image"
	_ "image/jpeg"
	qrcode "github.com/yeqown/go-qrcode/v2"
	standard "github.com/yeqown/go-qrcode/writer/standard"

	gozxing "github.com/makiuchi-d/gozxing"
	gozxing_qrcode "github.com/makiuchi-d/gozxing/qrcode"
)

// https://github.com/yeqown/go-qrcode/blob/f42371933e21f873d58fa712b88c37648cd489b8/writer/standard/image_option.go
// https://github.com/yeqown/go-qrcode/blob/3dec3839f3d3487bda5dd6cc518ba8ce2212692c/writer/standard/README.md
// https://github.com/yeqown/go-qrcode/blob/5de5083702af659f1fe10a94c90065d239bbd920/qrcode.go#L21
// https://github.com/yeqown/go-qrcode/blob/f42371933e21f873d58fa712b88c37648cd489b8/writer/standard/writer.go#L207
// https://github.com/ciaochaos/qrbtf
func Generate( url string , output_path string ) {
	// 1.) Make the QR Code Object
	qrc , _ := qrcode.New( url )

	// who knows , we can't get it to work
	// 2023/01/28 18:47:31 w=740, h=740, logoW=300, logoH=300, logo is over than 1/5 of QRCode
	// options := []standard.ImageOption{
	// 	standard.WithLogoImageFilePNG( "logo.png" ) ,
	// }
	// writer , _ := standard.New( "test.png" , options... )

	// 2.) Write QR Code Object to File
	writer , _ := standard.New( output_path )
	qrc.Save( writer )
}

func Decode( file_path string ) {
	file , _ := os.Open( file_path )
	img , _ , _ := image.Decode( file )
	// prepare BinaryBitmap
	bmp , _ := gozxing.NewBinaryBitmapFromImage( img )
	// decode image
	qrReader := gozxing_qrcode.NewQRCodeReader()
	result , _ := qrReader.Decode( bmp , nil )
	fmt.Println( result )
}