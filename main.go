package main

import (
	"flag"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
)

type TlsConfig struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

type HttpConfig struct {
	TlsConfig TlsConfig `yaml:"tls_config"`
}

type EmailConfig struct {
	SendResolved bool   `yaml:"send_resolved"`
	From         string `yaml:"from,omitempty"`
	To           string `yaml:"to"`
}

type WebhookConfig struct {
	Url          string     `yaml:"url"`
	SendResolved bool       `yaml:"send_resolved"`
	HttpConfig   HttpConfig `yaml:"http_config"`
}

type Receiver struct {
	Name           string          `yaml:"name"`
	EmailConfigs   []EmailConfig   `yaml:"email_configs,omitempty"`
	WebhookConfigs []WebhookConfig `yaml:"webhook_configs,omitempty"`
}

type Receivers struct {
	Receivers []Receiver `yaml:"receivers"`
}

type Route struct {
	Matchers []string `yaml:"matchers"`
	Receiver string   `yaml:"receiver"`
}

type Routes struct {
	Routes []Route `yaml:"routes"`
}

type RootRoute struct {
	GroupBy        []string `yaml:"group_by"`
	GroupWait      string   `yaml:"group_wait,omitempty"`
	GroupInterval  string   `yaml:"group_interval,omitempty"`
	RepeatInterval string   `yaml:"repeat_interval,omitempty"`
	Receiver       string   `yaml:"receiver"`
	Routes         []Route  `yaml:"routes"`
}

type Service struct {
	Owner              string `yaml:"owner"`
	ContactEmail       string `yaml:"contact_email,omitempty"`
	ContactChat        string `yaml:"contact_chat,omitempty"`
	AlertEmail         string `yaml:"alert_email,omitempty"`
	WebhookUrl         string `yaml:"webhook_url,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify,omitempty"`
}

type Services struct {
	Services map[string]Service `yaml:"services"`
}

func main() {

	servicesFile := "./services.yaml"
	routesDirectory := "./routes"
	alertReceiversFile := "./am_receivers.yaml"
	alertRoutesFile := "./am_routes.yaml"
	orphanAlertEmail := "unassigned-alerts@example.com"

	var services Services
	var receivers Receivers
	var routes Routes

	flag.StringVar(&servicesFile, "service-data", servicesFile, "Service Data")
	flag.StringVar(&routesDirectory, "routes-directory", routesDirectory, "AM Routing confguration directory")
	flag.StringVar(&alertRoutesFile, "alert-routing-file", alertRoutesFile, "Alert Routes Configuration File")
	flag.StringVar(&alertReceiversFile, "alert-receivers-file", alertReceiversFile, "Alert Receivers Configuration File")
	flag.StringVar(&orphanAlertEmail, "orphanAlertEmail", orphanAlertEmail, "Unowned Alert Receiver Email Address")
	flag.Parse()

	services.getConf(servicesFile)

	receivers.generateReceivers(orphanAlertEmail, services)

	routes.generateRoutes(services)

	if err := generateReceiversFile(routesDirectory, alertReceiversFile, &receivers); err != nil {
		log.Fatal(err)
		return
	}
	if err := generateRoutesFile(routesDirectory, alertRoutesFile, &routes); err != nil {
		log.Fatal(err)
		return
	}
}

func getUrl(c string) string {
	//TODO
	return (c)
}

func (r *Receivers) generateReceivers(orphanAlertEmail string, services Services) *Receivers {

	for service, serviceData := range services.Services {
		if len(serviceData.AlertEmail) > 0 {

			r.Receivers = append(r.Receivers, Receiver{
				Name: "email:" + service,
				EmailConfigs: []EmailConfig{
					EmailConfig{
						SendResolved: true,
						To:           serviceData.AlertEmail}}})
		}
		if len(serviceData.WebhookUrl) > 0 {
			r.Receivers = append(r.Receivers, Receiver{
				Name: "webhook:" + service,
				WebhookConfigs: []WebhookConfig{
					WebhookConfig{
						Url: getUrl(
							serviceData.WebhookUrl),
						SendResolved: true,
						HttpConfig: HttpConfig{
							TlsConfig: TlsConfig{
								InsecureSkipVerify: serviceData.InsecureSkipVerify}}}}})
		}
	}

	r.Receivers = append(r.Receivers, Receiver{Name: "email:default-receiver",
		EmailConfigs: []EmailConfig{
			EmailConfig{
				SendResolved: true,
				To:           orphanAlertEmail}}})
	return r
}

func (r *Routes) generateRoutes(services Services) *Routes {

	for service, serviceData := range services.Services {
		if len(serviceData.AlertEmail) > 0 {
			r.Routes = append(r.Routes, Route{
				Matchers: []string{"service=\"" + service + "\""},
				Receiver: "email:" + service})
		}
		if len(serviceData.WebhookUrl) > 0 {
			r.Routes = append(r.Routes, Route{
				Matchers: []string{"service=\"" + service + "\""},
				Receiver: "webhook:" + service})
		}

	}
	return r
}

func (s *Services) getConf(servicesFile string) *Services {

	yamlFile, err := os.ReadFile(servicesFile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	if err := yaml.Unmarshal(yamlFile, s); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return s
}

func generateReceiversFile(outputDir string, filename string, receivers *Receivers) error {
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return err
	}

	alertConfig, err := yaml.Marshal(receivers)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(outputDir, filename), alertConfig, 0666); err != nil {
		return err
	}

	return nil
}

func generateRoutesFile(outputDir string, filename string, routes *Routes) error {
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return err
	}
	alertConfig, err := yaml.Marshal(map[string]RootRoute{
		"route": RootRoute{
			GroupBy:        []string{"alertname"},
			GroupWait:      "30s",
			GroupInterval:  "5m",
			RepeatInterval: "1h",
			Receiver:       "email:default-receiver",
			Routes:         routes.Routes}})

	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(outputDir, filename), alertConfig, 0666); err != nil {
		return err
	}

	return nil
}
