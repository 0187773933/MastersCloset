<center><h1>Windows Printer Install</h1></center>

1. [Download](https://raw.githubusercontent.com/0187773933/MastersCloset/master/misc/barcode%20printer%20drivers/4BARCODE.driver.6.0.0.40(windows).exe) and Install Barcode Printer Driver
2. Name it `_4BARCODE_4B_2054N`
3. Printer Properties ➡️ Preferences ➡️ Print Settings ➡️ Print Speed ➡️ 3.00 in/s ?

## Printing helper

Windows prints through `SumatraPDF.exe`. Keep it **next to `mct.exe`** — `build.sh`
copies it there automatically, and a copy also lives in this repo at
`misc/SumatraPDF.exe`.

If it ends up somewhere else, the app also looks in `%USERPROFILE%\.config\mct`, the
folder it was launched from, and `%PATH%` — or set **Settings ➡️ Printer ➡️ SumatraPDF
Path** to point straight at it. When it cannot be found, the print error lists every
location that was searched.
