package adminroutes

import (
	"os"
	"fmt"
	"bytes"
	"path/filepath"
	"encoding/base64"
	"mime/multipart"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	utils "github.com/0187773933/MastersCloset/v1/utils"
	types "github.com/0187773933/MastersCloset/v1/types"
	printer "github.com/0187773933/MastersCloset/v1/printer"
)

// weak attempt at sanitizing form input to build a "username"
func SanitizeUsername( first_name string , last_name string ) ( username string ) {
	if first_name == "" { first_name = "Not Provided" }
	if last_name == "" { last_name = "Not Provided" }
	sanitized_first_name := utils.SanitizeInputString( first_name )
	sanitized_last_name := utils.SanitizeInputString( last_name )
	username = fmt.Sprintf( "%s-%s" , sanitized_first_name , sanitized_last_name )
	return
}

func serve_failed_attempt( context *fiber.Ctx ) ( error ) {
	// context.Set( "Content-Type" , "text/html" )
	// return context.SendString( "<h1>no</h1>" )
	return context.SendFile( "./v1/server/html/admin_login.html" )
}

func ServeLoginPage( context *fiber.Ctx ) ( error ) {
	return context.SendFile( "./v1/server/html/admin_login.html" )
}

func ServeAuthenticatedPage( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	x_path := context.Route().Path
	url_key := strings.Split( x_path , "/admin" )
	if len( url_key ) < 2 { return context.SendFile( "./v1/server/html/admin_login.html" ) }
	// fmt.Println( "Sending -->" , url_key[ 1 ] , x_path )
	context.Set( "Cache-Control" , "public, max-age=1" );
	return context.SendFile( ui_html_pages[ url_key[ 1 ] ] )
}

func PrintTest( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	printer.PrintTicket( GlobalConfig.Printer , printer.PrintJob{
		FamilySize: 5 ,
		TotalClothingItems: 23 ,
		Shoes: 1 ,
		Accessories: 2 ,
		Seasonal: 1 ,
		FamilyName: "Cerbus" ,
	})
	return context.JSON( fiber.Map{
		"route": "/admin/print-test" ,
		"result": "success" ,
	})
}

func Print( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	var print_job printer.PrintJob
	json.Unmarshal( []byte( context.Body() ) , &print_job )
	fmt.Println( print_job )
	printer.PrintTicket( GlobalConfig.Printer , print_job )
	return context.JSON( fiber.Map{
		"route": "/admin/print" ,
		"result": true ,
	})
}

func PrintTwo( context *fiber.Ctx ) ( error ) {
	fmt.Println( "PrintTwo()" )
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	var print_job printer.PrintJob
	json.Unmarshal( []byte( context.Body() ) , &print_job )
	fmt.Println( print_job )
	printer.PrintTicket2( GlobalConfig.Printer , print_job )
	return context.JSON( fiber.Map{
		"route": "/admin/print2" ,
		"result": true ,
	})
}

func getAIParsedJSONOfAudioDescriptionOfFamily(audioTranscription string) (map[string]interface{}, error) {
	assistantInstructions := `
		convert the input into a structured JSON object, including only the details provided:
		{
			"self": {
				"age": ${AGE} ,
				"sex": ${SEX} ,
				"first_name": "${FIRST_NAME}" ,
				"middle_name": "${MIDDLE_NAME} ,
				"last_name": "${LAST_NAME}" ,
				"email_address": "${EMAIL_ADDRESS}" ,
				"phone_number": "${PHONE_NUMBER}" ,
				"street_number": "${STREET_NUMBER}" ,
				"street_name": "${STREET_NAME}" ,
				"city": "${CITY}" ,
				"state": "${STATE}" ,
				"zip_code": "${ZIP_CODE}" ,
				"birth_day": "${BIRTH_DAY}" ,
				"birth_month": "${BIRTH_MONTH}" ,
				"birth_year": "${BIRTH_YEAR}"
			} ,
			"adults": [
				{
					"age": ${ADULT_1_AGE} ,
					"sex": ${ADULT_1_SEX} ,
					"spouse": ${ADULT_1_SPOUSE_OF_SELF}
				} ,
				...
			]
			"children": [
				{
					"age": ${CHILD_1_AGE} ,
					"sex": ${CHILD_1_SEX}
				} ,
				...
			]
		}
	`
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", GlobalConfig.OpenAIAPIKey),
	}
	data := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "system", "content": assistantInstructions},
			{"role": "user", "content": audioTranscription},
		},
	}
	// fmt.Println( data )
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Error: %s", body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	fmt.Println( result )
	choices := result["choices"].([]interface{})
	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)
	// fmt.Println( content )
	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonResult); err != nil {
		return nil, err
	}
	fmt.Println( jsonResult )
	return jsonResult, nil
}

func AudioToBaseUserStructure( c *fiber.Ctx ) ( error ) {
	fmt.Println( "AudioToBaseUserStructure" )
	var data types.AudioData
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"result": false, "error": err.Error()})
	}

	audioData, err := base64.StdEncoding.DecodeString(data.Audio)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"result": false, "error": "invalid base64 audio"})
	}

	primaryMimeType := strings.Split(data.Type, ";")[0]
	fmt.Println( primaryMimeType )
	fileSuffix := map[string]string{
		"audio/ogg":  ".ogg",
		"audio/webm": ".webm",
		"audio/wav":  ".wav",
		"audio/mpeg": ".mp3",
	}[primaryMimeType]

	if fileSuffix == "" {
		fileSuffix = ".tmp"
	}

	tmpFile, err := ioutil.TempFile("", "*"+fileSuffix)
	tmpFilePath := tmpFile.Name()
	fmt.Println( tmpFilePath )
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"result": false, "error": "could not create temp file"})
	}
	defer os.Remove(tmpFilePath)

	if _, err := tmpFile.Write(audioData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"result": false, "error": "could not write to temp file"})
	}

	// fmt.Println( "stage 2" )

	url := "https://api.openai.com/v1/audio/transcriptions"
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", GlobalConfig.OpenAIAPIKey),
	}

	file, err := os.Open(tmpFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"result": false, "error": "could not open temp file"})
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(tmpFilePath))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"result": false, "error": "could not create form file"})
	}
	if _, err := io.Copy(part, file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"result": false, "error": "could not copy file content"})
	}
	writer.WriteField("model", "whisper-1")
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"result": false, "error": "could not create request"})
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"result": false, "error": "request failed"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).JSON(fiber.Map{"result": false, "error": string(body)})
	}

	var decoded map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"result": false, "error": "could not decode response"})
	}

	// fmt.Println( "stage 3" )

	if _, ok := decoded["text"]; !ok {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"result": false})
	}
	instructions := decoded["text"].(string)
	fmt.Println( instructions )
	aiParsedJSON, err := getAIParsedJSONOfAudioDescriptionOfFamily(instructions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"result": false, "error": err.Error() , "instructions": instructions , "decoded": aiParsedJSON})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"result": aiParsedJSON})
}