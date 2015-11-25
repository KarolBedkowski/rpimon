package monitor

import (
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/cfg"
	//	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/logging"
	"net/http"
	"regexp"
	"strconv"
)

var decoder = schema.NewDecoder()

type confForm cfg.MonitorConfiguration

type (
	confPageContext struct {
		*context.BasePageContext
		Form   *confForm
		New    bool
		Errors []string
	}
)

// Rules for validate monitored host address
var (
	reValidateTcpAddress  = regexp.MustCompile(`^[\w.-]{1,63}:\d{1,5}$`)
	reValidateHTTPAddress = regexp.MustCompile(`^(https?://)?[\w.-]{1,63}(:\d{1,5})?(/.*)?$`)
	reValidatePingAddress = regexp.MustCompile(`^[\w.-]{1,63}$`)
)

func (f *confForm) validate() (errors []string) {
	if f.UpdateInterval < 0 {
		errors = append(errors, "Invalud update interval")
	}
	if f.LoadWarning < 0 {
		errors = append(errors, "Invalid load warning level")
	}
	if f.LoadError < 0 {
		errors = append(errors, "Invalid load error level")
	}
	if f.LoadWarning > f.LoadError && f.LoadError > 0 {
		errors = append(errors, "Load warning level should be lower than error level")
	}
	if f.RAMUsageWarning < 0 || f.RAMUsageWarning > 100 {
		errors = append(errors, "Invalid RAM usage warning level - should be in range 0-100")
	}
	if f.SwapUsageWarning < 0 || f.SwapUsageWarning > 100 {
		errors = append(errors, "Invalid Swap usage warning level - should be in range 0-100")
	}
	if f.DefaultFSUsageWarning < 0 || f.DefaultFSUsageWarning > 100 {
		errors = append(errors, "Invalid FS usage warning level - should be in range 0-100")
	}
	if f.DefaultFSUsageError < 0 || f.DefaultFSUsageError > 100 {
		errors = append(errors, "Invalid FS usage Error level - should be in range 0-100")
	}
	if f.DefaultFSUsageError < f.DefaultFSUsageWarning && f.DefaultFSUsageError > 0 {
		errors = append(errors, "FS usage Error level should be higher than warning level")
	}
	if f.CPUTempWarning < 0 {
		errors = append(errors, "Invalid CPU temperature warning level")
	}
	if f.CPUTempError < 0 {
		errors = append(errors, "Invalid CPU temperature error level")
	}
	if f.CPUTempError < f.CPUTempWarning && f.CPUTempError > 0 {
		errors = append(errors, "CPU temperature error level should be higher than warning level")
	}

	// validate monitored hosts
	for idx, h := range f.MonitoredHosts {
		hostName := h.Name
		if h.Name == "" {
			hostName = strconv.Itoa(idx + 1)
		}
		if h.Address == "" {
			errors = append(errors, "Missing address for host "+hostName)
		} else {
			validAddress := true
			switch h.Method {
			case "tcp":
				validAddress = reValidateTcpAddress.MatchString(h.Address)
			case "http":
				validAddress = reValidateHTTPAddress.MatchString(h.Address)
			case "ping":
				validAddress = reValidatePingAddress.MatchString(h.Address)
			default:
				errors = append(errors, "Invalid method")
				l.Warn("MonitorConfiguration: Validation: invalid method: %#v", h)
			}
			if !validAddress {
				errors = append(errors, "Invalid address for host "+hostName)
			}
			if h.Interval < 0 || h.Interval > 9999 {
				errors = append(errors, "Invalid Interval for host "+hostName)
			}
		}
	}

	// validate monitored services
	for idx, s := range f.MonitoredServices {
		if s.Port < 1 || s.Port > 65535 {
			if s.Name != "" {
				errors = append(errors, "Invalid port for service "+strconv.Itoa((idx+1)))
			} else {
				errors = append(errors, "Invalid port for service '"+s.Name+"'")
			}
		}
	}

	return
}

// cleanup data in form - ie. empty rows in lists
func (f *confForm) cleanup() {
	// cleanup monitored services
	var servs []cfg.MonitoredService
	for _, serv := range f.MonitoredServices {
		if serv.Port > 0 {
			if serv.Name == "" {
				serv.Name = "Connection to port " + strconv.Itoa(int(serv.Port))
			}
			servs = append(servs, serv)
		}
	}
	f.MonitoredServices = servs
	// cleanup monitored shosts
	var hosts []cfg.MonitoredHost
	for _, host := range f.MonitoredHosts {
		if host.Address != "" {
			if host.Name == "" {
				host.Name = "Connection to " + host.Address
			}
			hosts = append(hosts, host)
		}
	}
	f.MonitoredHosts = hosts
}

func confPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	form := confForm{}
	form = confForm(*cfg.Configuration.Monitor)
	ctx := &confPageContext{BasePageContext: bctx,
		Form: &form,
	}

	switch r.Method {
	case "POST":
		r.ParseForm()
		// remove monitored services - fill in only with new data
		form.MonitoredServices = nil
		form.MonitoredHosts = nil
		if err := decoder.Decode(ctx.Form, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
		}
		errors := ctx.Form.validate()
		if errors == nil || len(errors) == 0 {
			form.cleanup()
			*cfg.Configuration.Monitor = cfg.MonitorConfiguration(form)
			err := cfg.SaveConfiguration()
			if err != nil {
				ctx.AddFlashMessage("Saving configuration error: "+err.Error(),
					"error")
			} else {
				ctx.AddFlashMessage("Configuration saved.", "success")
			}
			ctx.Save()
			http.Redirect(w, r, r.URL.String(), http.StatusFound)
			return
		}
		ctx.Errors = errors
		ctx.AddFlashMessage("Validation errors!", "error")
	case "GET":
	}
	ctx.Save()
	app.RenderTemplateStd(w, ctx, "monitor/monitor-conf.tmpl")
}
