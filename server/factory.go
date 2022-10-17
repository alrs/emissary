package server

import (
	"embed"
	"fmt"
	"html/template"
	"strings"
	"sync"

	"github.com/EmissarySocial/emissary/config"
	"github.com/EmissarySocial/emissary/domain"
	"github.com/EmissarySocial/emissary/queue"
	"github.com/EmissarySocial/emissary/render"
	"github.com/EmissarySocial/emissary/service"
	"github.com/benpate/derp"
	"github.com/benpate/icon"
	"github.com/benpate/nebula"
	"github.com/benpate/steranko"
	"github.com/labstack/echo/v4"
	"github.com/spf13/afero"
)

// Factory manages all server-level services, and generates individual
// domain factories for each domain
type Factory struct {
	storage config.Storage
	config  config.Config
	mutex   sync.RWMutex

	// Server-level services
	layoutService   service.Layout
	templateService service.Template
	externalService service.External
	emailService    service.ServerEmail
	taskQueue       *queue.Queue
	embeddedFiles   embed.FS

	attachmentOriginals afero.Fs
	attachmentCache     afero.Fs

	// Widget Libraries
	contentLibrary nebula.Library

	domains map[string]*domain.Factory
}

// NewFactory uses the provided configuration data to generate a new Factory
// if there are any errors connecting to a domain's datasource, NewFactory will derp.Report
// the error, but will continue loading without those domains.
func NewFactory(storage config.Storage, embeddedFiles embed.FS) *Factory {

	factory := Factory{
		storage:        storage,
		mutex:          sync.RWMutex{},
		domains:        make(map[string]*domain.Factory, 0),
		embeddedFiles:  embeddedFiles,
		contentLibrary: nebula.NewLibrary(),
		taskQueue:      queue.NewQueue(128, 16),
	}

	// Global Layout service
	factory.layoutService = service.NewLayout(
		factory.Filesystem(),
		factory.FuncMap(),
		[]config.Folder{},
	)

	// Global Template Service
	factory.templateService = *service.NewTemplate(
		factory.Layout(),
		factory.Filesystem(),
		factory.FuncMap(),
		[]config.Folder{},
	)

	factory.externalService = service.NewExternal(factory.config.Providers)

	factory.emailService = service.NewServerEmail(
		factory.Filesystem(),
		factory.FuncMap(),
		[]config.Folder{},
	)

	go factory.start()

	return &factory
}

func (factory *Factory) start() {

	fmt.Println("Factory: Waiting for configuration file...")

	filesystemService := factory.Filesystem()

	// Read configuration files from the channel
	for config := range factory.storage.Subscribe() {

		fmt.Println("Factory: received new configuration...")

		if attachmentOriginals, err := filesystemService.GetAfero(config.AttachmentOriginals); err == nil {
			factory.attachmentOriginals = attachmentOriginals
		} else {
			derp.Report(derp.Wrap(err, "server.Factory.start", "Error getting attachment original directory", config.AttachmentOriginals))
		}

		if attachmentCache, err := filesystemService.GetAfero(config.AttachmentCache); err == nil {
			factory.attachmentCache = attachmentCache
		} else {
			derp.Report(derp.Wrap(err, "server.Factory.start", "Error getting attachment cache directory", config.AttachmentCache))
		}

		factory.config = config

		// Mark all domains for deletion (then unmark them later)
		for index := range factory.domains {
			factory.domains[index].MarkForDeletion = true
		}

		// Refresh cached values in global services
		factory.layoutService.Refresh(config.Layouts)
		factory.templateService.Refresh(config.Templates)
		factory.emailService.Refresh(config.Emails)
		factory.externalService.Refresh(config.Providers)

		// Insert/Update a factory for each domain in the configuration
		for _, domainConfig := range config.Domains {

			// Try to find the domain
			if existing := factory.domains[domainConfig.Hostname]; existing != nil {

				if err := existing.Refresh(domainConfig, config.Providers, factory.attachmentOriginals, factory.attachmentCache); err != nil {
					derp.Report(derp.Wrap(err, "server.Factory.start", "Error refreshing domain", domainConfig.Hostname))
				}

				// Even if there's an error "refreshing" the domain, we don't want to delete it
				existing.MarkForDeletion = false
				continue
			}

			// Fall through means that the domain does not exist, so we need to create it
			newDomain, err := domain.NewFactory(
				domainConfig,
				config.Providers,
				&factory.emailService,
				&factory.layoutService,
				&factory.templateService,
				&factory.externalService,
				&factory.contentLibrary,
				factory.taskQueue,
				factory.attachmentOriginals,
				factory.attachmentCache,
			)

			if err != nil {
				derp.Report(derp.Wrap(err, "server.Factory.start", "Unable to start domain", domainConfig))
				continue
			}

			// If there are no errors, then add the domain to the list.
			factory.domains[newDomain.Hostname()] = newDomain
		}

		// Any domains that are still marked for deletion will be gracefully closed, then removed
		for domainID := range factory.domains {
			if factory.domains[domainID].MarkForDeletion {
				factory.domains[domainID].Close()
				delete(factory.domains, domainID)
			}
		}
	}
}

/****************************
 * Server Config Methods
 ****************************/

// Config returns the current configuration for the Factory
func (factory *Factory) Config() config.Config {

	// Read lock the mutex
	factory.mutex.RLock()
	defer factory.mutex.RUnlock()

	result := factory.config
	return result
}

// UpdateConfig updates the configuration for the Factory
func (factory *Factory) UpdateConfig(value config.Config) error {

	// Write lock the mutex
	factory.mutex.Lock()
	defer factory.mutex.Unlock()

	factory.config = value

	if err := factory.storage.Write(value); err != nil {
		return derp.Wrap(err, "server.Factory.UpdateConfig", "Error writing configuration", value)
	}

	return nil
}

/****************************
 * Domain Methods
 ****************************/

// ListDomains returns a list of all domains in the Factory
func (factory *Factory) ListDomains() []config.Domain {
	return factory.config.Domains
}

// PutDomain adds a domain to the Factory
func (factory *Factory) PutDomain(domain config.Domain) error {

	factory.mutex.Lock()
	defer factory.mutex.Unlock()

	// Add the domain to the collection
	factory.config.Domains.Put(domain)

	// Try to write the configuration to the storage service
	if err := factory.storage.Write(factory.config); err != nil {
		return derp.Wrap(err, "server.Factory.WriteConfig", "Error writing configuration")
	}

	// The storage service will trigger a new configuration via the Subscrbe() channel

	return nil
}

// DomainByID finds a domain in the configuration by its ID
func (factory *Factory) DomainByID(domainID string) (config.Domain, error) {

	factory.mutex.RLock()
	defer factory.mutex.RUnlock()

	// If "new" then create a new domain
	if strings.ToLower(domainID) == "new" {
		return config.NewDomain(), nil
	}

	// Search for the domain in the configuration
	if domain, ok := factory.config.Domains.Get(domainID); ok {
		return domain, nil
	}

	// Not found, so return an error
	return config.NewDomain(), derp.NewNotFoundError("server.Factory.DomainByID", "DomainID not found", domainID)
}

// DeleteDomain removes a domain from the Factory
func (factory *Factory) DeleteDomain(domainID string) error {

	factory.mutex.Lock()
	defer factory.mutex.Unlock()

	// Delete the domain from the collection
	factory.config.Domains.Delete(domainID)

	// Write changes to the storage engine.
	if err := factory.storage.Write(factory.config); err != nil {
		return derp.Wrap(err, "server.Factory.DeleteDomain", "Error saving configuration")
	}

	return nil
}

/****************************
 * OAuth Connection Methods
 ****************************/

// PutConnection adds a domain to the Factory
func (factory *Factory) PutProvider(oauthClient config.Provider) error {

	factory.mutex.Lock()
	defer factory.mutex.Unlock()

	// Add the domain to the collection
	factory.config.Providers.Put(oauthClient)

	// Try to write the configuration to the storage service
	if err := factory.storage.Write(factory.config); err != nil {
		return derp.Wrap(err, "server.Factory.WriteConfig", "Error writing configuration")
	}

	// The storage service will trigger a new configuration via the Subscrbe() channel

	return nil
}

// DeleteConnection removes a domain from the Factory
func (factory *Factory) DeleteProvider(providerID string) error {

	factory.mutex.Lock()
	defer factory.mutex.Unlock()

	// Delete the connection from the collection
	factory.config.Providers.Delete(providerID)

	// Write changes to the storage engine.
	if err := factory.storage.Write(factory.config); err != nil {
		return derp.Wrap(err, "server.Factory.DeleteDomain", "Error saving configuration")
	}

	return nil
}

/****************************
 * Factory Methods
 ****************************/

// ByContext retrieves a domain using an echo.Context
func (factory *Factory) ByContext(ctx echo.Context) (*domain.Factory, error) {

	host := factory.NormalizeHostname(ctx.Request().Host)
	return factory.ByDomainName(host)
}

// ByDomainID retrieves a domain using a DomainID
func (factory *Factory) ByDomainID(domainID string) (config.Domain, *domain.Factory, error) {

	// Look up the domain name for this domainID
	domainConfig, err := factory.DomainByID(domainID)

	if err != nil {
		return config.Domain{}, nil, derp.Wrap(err, "server.Factory.ByDomainID", "Error finding domain configuration", domainID)
	}

	// Return the domain
	result, err := factory.ByDomainName(domainConfig.Hostname)

	if err != nil {
		return config.Domain{}, nil, derp.Wrap(err, "server.Factory.ByDomainID", "Error finding domain", domainConfig.Hostname)
	}

	return domainConfig, result, nil
}

// ByDomainName retrieves a domain using a Domain Name
func (factory *Factory) ByDomainName(name string) (*domain.Factory, error) {

	factory.mutex.RLock()
	defer factory.mutex.RUnlock()

	if domain, ok := factory.domains[name]; ok {
		return domain, nil
	}

	return nil, derp.New(404, "factory.ByDomainName.Get", "Unrecognized domain name", name)
}

// NormalizeHostname removes some inconsistencies in host names, including a leading "www", if present
func (factory *Factory) NormalizeHostname(hostname string) string {

	hostname = strings.ToLower(hostname)

	if dotIndex := strings.Index(hostname, "."); dotIndex > 0 {

		if subdomain := hostname[0 : dotIndex-1]; subdomain == "www" {
			return hostname[dotIndex+1:]
		}
	}

	return hostname
}

/****************************
 * Other Global Services
 ****************************/

// Layout returns the global layout service
func (factory *Factory) Layout() *service.Layout {
	return &factory.layoutService
}

// FuncMap returns the global funcMap (used by all templates)
func (factory *Factory) FuncMap() template.FuncMap {
	return render.FuncMap(factory.Icons())
}

// Icons returns the global icon collection
func (factory *Factory) Icons() icon.Provider {
	return service.Icons{}
}

// Filesystem returns the global filesystem service
func (factory *Factory) Filesystem() service.Filesystem {
	return service.NewFilesystem(factory.embeddedFiles)
}

// Email returns the global email service
func (factory *Factory) Email() *service.ServerEmail {
	return &factory.emailService
}

// Steranko implements the steranko.Factory method, used for locating the specific
// steranko instance used by a domain.
func (factory *Factory) Steranko(ctx echo.Context) (*steranko.Steranko, error) {

	result, err := factory.ByContext(ctx)

	if err != nil {
		return nil, derp.Wrap(err, "server.Factory.Steranko", "Unable to locate factory for domain", ctx.Request().Host)
	}

	return result.Steranko(), nil
}
