package main

import (
	"GoUnder/cmd"
	"fmt"
)

func main() {
	fmt.Print(`
   ██████╗  ██████╗  ██╗   ██╗██╗   ██╗███╗   ██╗██████╗ ███████╗██████╗ 
  ██╔════╝ ██╔═══██╗ ██║   ██║██║   ██║████╗  ██║██╔══██╗██╔════╝██╔══██╗
  ██║  ███╗██║   ██║ ██║   ██║██║   ██║██╔██╗ ██║██║  ██║█████╗  ██████╔╝
  ██║   ██║██║   ██║ ╚██╗ ██╔╝██║   ██║██║╚██╗██║██║  ██║██╔══╝  ██╔══██╗
  ╚██████╔╝╚██████╔╝  ╚████╔╝ ╚██████╔╝██║ ╚████║██████╔╝███████╗██║  ██║
   ╚═════╝  ╚═════╝    ╚═══╝   ╚═════╝ ╚═╝  ╚═══╝╚═════╝ ╚══════╝╚═╝  ╚═╝
                 GoUnder v0.1 - CDN Bypass & Fingerprint Scan

      ⛓ Engine     : GoLang Fast Recon Core
      🎯 Target     : Reveal Real IP Behind CDN
      🔍 Fingerprint: CMS / Framework / Server Detection
      ☠ Status     : Cloak Engaged - Tracing Origin...
`)
	cmd.Execute()
}
