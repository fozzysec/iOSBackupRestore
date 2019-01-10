package main

import (
    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
    "os"
    "io"
)
var DB_FILE = "Manifest.db"
//var WECHAT_ID = "AppDomain-com.tencent.xin"

func main(){
    if len(os.Args) < 3 {
        fmt.Printf("usage: %s export source_dir export_dir AppID\n", os.Args[0])
        fmt.Printf("example: %s Backup/000000000000000000 OutPut AppDomain-com.tencent.xin\n")
        return
    }
    operation := os.Args[1]
    sourceDir := os.Args[2]
    if operation == "export" {
        destDir := os.Args[3]
        AppID := os.Args[4]
        export(sourceDir, destDir, AppID)
    } else if operation == "query"{
        getAppIDList(sourceDir)
    } else {
        fmt.Println("Bad operation.")
    }
}

func export(sourceDir string, destDir string, AppID string){
    FileLists := make(map[string]string)
    DirLists := make(map[string]string)
    getList(AppID, sourceDir, FileLists, DirLists)
    makeDirs(destDir, DirLists)
    PathLists := make(map[string]string)
    getPathPrefix(FileLists, PathLists)
    copyFiles(sourceDir, destDir, PathLists)
}

func onFailExit(err error){
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
}
func onFailWarning(err error){
    if err != nil {
        fmt.Println(err.Error())
    }
}
func getAppIDList(dir string){
    dbPath := fmt.Sprintf("%s/%s", dir, DB_FILE)
    db, err := sql.Open("sqlite3", dbPath)
    onFailExit(err)
    defer db.Close()
    sql := "SELECT DISTINCT domain FROM Files ORDER BY domain ASC"
    rows, err := db.Query(sql)
    defer rows.Close()
    onFailExit(err)
    for rows.Next(){
        var AppID string
        err = rows.Scan(&AppID)
        onFailExit(err)
        fmt.Println(AppID)
    }
    err = rows.Err()
    onFailExit(err)
}

func getList(AppID string, dir string, FileLists map[string]string, DirLists map[string]string){
    dbPath := fmt.Sprintf("%s/%s", dir, DB_FILE)
    db, err := sql.Open("sqlite3", dbPath)
    onFailExit(err)
    defer db.Close()
    sqlstatement := "SELECT fileID, relativePath, flags FROM Files WHERE domain = ? ORDER BY flags DESC"
    stmt, err := db.Prepare(sqlstatement)
    onFailExit(err)
    defer stmt.Close()
    rows, err := stmt.Query(AppID)
    onFailExit(err)
    defer rows.Close()
    for rows.Next(){
        var FileID, Path string
        var flag int
        err = rows.Scan(&FileID, &Path, &flag)
        onFailExit(err)
        if flag == 2 {
            DirLists[FileID] = Path
        } else {
            FileLists[FileID] = Path
        }
    }
    err = rows.Err()
    onFailExit(err)
}

func makeDirs(destDir string, DirLists map[string]string){
    for _, Path := range DirLists {
        dstPath := fmt.Sprintf("%s/%s", destDir, Path)
        os.MkdirAll(dstPath, 0755)
    }
}

func getPathPrefix(FileLists map[string]string, PathLists map[string]string){
    for srcPath, destPath := range FileLists {
        PathLists[srcPath[:2]+"/"+srcPath] = destPath
    }
}

func copyFiles(srcDir string, destDir string, FileLists map[string]string){
    for srcFilePath, destFilePath := range FileLists {
        srcFilePath = fmt.Sprintf("%s/%s", srcDir, srcFilePath)
        destFilePath = fmt.Sprintf("%s/%s", destDir, destFilePath)
        srcFile, err := os.Open(srcFilePath)
        if err != nil {
            fmt.Println(err.Error())
            continue
        }
        destFile, err := os.Create(destFilePath)
        if err != nil {
            fmt.Println(err.Error())
            continue
        }
        _, err = io.Copy(destFile, srcFile)
        if err != nil {
            fmt.Println(err.Error())
            continue
        }
        err = destFile.Sync()
        if err != nil {
            fmt.Println(err.Error())
            continue
        }
        defer srcFile.Close()
        defer destFile.Close()
    }
}
