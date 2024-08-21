package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

type Ticket struct {
	Title       string `yaml:"title"`
	Type        string `yaml:"type"`
	Area        string `yaml:"area"`
	AssignedTo  string `yaml:"assignedTo,omitempty"`
	Iteration   string `yaml:"iteration"`
	Org         string `yaml:"org"`
	Project     string `yaml:"project"`
	Description string `yaml:"description,omitempty"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	var content bytes.Buffer

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		content.WriteString(input)
	}

	var ticket Ticket
	err := yaml.Unmarshal(content.Bytes(), &ticket)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return
	}

	createTicket(ticket)
}

func escapeString(str string) string {
	str = strings.ReplaceAll(str, `"`, `\"`)
	str = strings.ReplaceAll(str, "\n", "\\n")
	return str
}

func sanitizeParams(ticket Ticket) Ticket {
	ticket.Title = escapeString(ticket.Title)
	ticket.Type = escapeString(ticket.Type)
	ticket.Area = escapeString(ticket.Area)
	ticket.AssignedTo = escapeString(ticket.AssignedTo)
	ticket.Iteration = escapeString(ticket.Iteration)
	ticket.Org = escapeString(ticket.Org)
	ticket.Project = escapeString(ticket.Project)
	return ticket
}

func optionallyAssign(assignedTo string) string {
	if assignedTo != "" {
		return fmt.Sprintf(`--assigned-to "%s"`, assignedTo)
	}
	return ""
}

func optionallyDescribe(description string, ttype string) string {
	if description != "" {
		switch ttype {
		case "Product Backlog Item":
			return fmt.Sprintf(`-f "Description=%s"`, description)
		case "Bug":
			return fmt.Sprintf(`-f "Repro Steps=%s"`, description)
		}
	}
	return ""
}

func generateCmd(ticket Ticket) string {
	cmd := fmt.Sprintf(`az boards work-item create --title "%s" --type "%s" --area "%s" %s --iteration "%s" --org "%s" --project "%s" %s --open`,
		ticket.Title,
		ticket.Type,
		ticket.Area,
		optionallyAssign(ticket.AssignedTo),
		ticket.Iteration,
		ticket.Org,
		ticket.Project,
		optionallyDescribe(ticket.Description, ticket.Type),
	)
	return cmd
}

func createTicket(ticket Ticket) {
	sanitizedTicket := sanitizeParams(ticket)
	cmd := generateCmd(sanitizedTicket)
	fmt.Println(cmd)
	execCmd := exec.Command("bash", "-c", cmd)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error executing command:", err)
		fmt.Println(string(output))
		return
	}
	fmt.Println(string(output))
}
