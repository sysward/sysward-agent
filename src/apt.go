package main

import (
  "os/exec"
  "strings"
  "strconv"
  "encoding/json"
  "fmt"
  "encoding/base64"
)

func updatePackage(pkg string) error {
  out, err := exec.Command("apt-get", "install", "-y", pkg).CombinedOutput()
  fmt.Println(string(out))
  return err
}

func getSourcesList() []Source {
  out, _ := exec.Command("grep", "-h", "^deb", "/etc/apt/sources.list", "/etc/apt/sources.list.d/*").Output()
  out_arr := strings.Split(strings.TrimSpace(string(out)), "\n")
  sources := make([]Source, len(out_arr))
  for index, o := range  out_arr {
    x := strings.Split(o, " ")
    src := false
    if x[0] == "deb-src" { src = true }
    sources[index] = Source{x[1], src, x[2:]}
  }
  return sources
}

func getChangelog(package_name string) string {
  changelog, _ := exec.Command("apt-get", "changelog", package_name).Output()
  changelog_encoded := base64.StdEncoding.EncodeToString(changelog)
  return changelog_encoded
}


func buildPackageList() []OsPackage {
  out, err := exec.Command("python", "trex.py").Output()
  if err != nil { panic(err) }
  var packages []OsPackage

  err = json.Unmarshal(out, &packages)
  if err != nil { panic(err) }

  return packages
}

func updatePackageLists() {
  out, err := exec.Command("apt-get", "update").Output()
  if err != nil {
    fmt.Println(string(out))
    panic(err)
  }
}

func updateCounts() Updates {
  out, err := exec.Command("/usr/lib/update-notifier/apt-check").CombinedOutput()
  if err != nil { panic(err) }
  output := strings.TrimSpace(string(out))
  ups := strings.Split(output, ";")
  regular, err := strconv.Atoi(ups[0])
  if err != nil { panic(err) }
  security, err := strconv.Atoi(ups[1])
  if err != nil { panic(err) }
  return Updates{regular, security}
}
