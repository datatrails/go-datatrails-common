package azblob

type GetMetadata int

const (
	NoMetadata GetMetadata = iota
	OnlyMetadata
	BothMetadataAndBlob
)

func (g GetMetadata) String() string {
	return [...]string{"No metadata handling", "Only metadata", "metadata and blob"}[g]
}

// StorerOptions - optional args for specifying optional behaviour
type StorerOptions struct {
	leaseID     string
	metadata    map[string]string
	tags        map[string]string
	getMetadata GetMetadata
	getTags     bool
	sizeLimit   int64
}

type Option func(*StorerOptions)

// Specifying an option that is no used is silently ignored. i.e. Specifying
// WithMetadata() in a call to Reader() will not raise an error.

// WithLeaseID - specifies LeaseId - Reader() and Write()
func WithLeaseID(leaseID string) Option {
	return func(a *StorerOptions) {
		a.leaseID = leaseID
	}
}

// WithMetadata specifies metadata to add - Write() only
func WithMetadata(metadata map[string]string) Option {
	return func(a *StorerOptions) {
		a.metadata = metadata
	}
}

// WithGetMetadata specifies to get metadata - Reader() only.
func WithGetMetadata(value GetMetadata) Option {
	return func(a *StorerOptions) {
		a.getMetadata = value
	}
}

// WithTags specifies tags to add - Reader() and Write(). For Write) the tags are written
// with the blob. For Reader() the tags are used to apply ownership permissions.
func WithTags(tags map[string]string) Option {
	return func(a *StorerOptions) {
		a.tags = tags
	}
}

func WithGetTags() Option {
	return func(a *StorerOptions) {
		a.getTags = true
	}
}

// WithSizeLimit specifies the size limit of the blob.
// -1 for unlimited. 0+ for limited.
func WithSizeLimit(sizeLimit int64) Option {
	return func(a *StorerOptions) {
		a.sizeLimit = sizeLimit
	}
}