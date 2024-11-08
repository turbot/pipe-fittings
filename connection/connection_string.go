package connection

// ConnectionString is a ConnectionStringProvider implementation that has a static connection string
type ConnectionString struct {
	ConnectionString string
}

func NewConnectionString(connectionString string) ConnectionStringProvider {
	return &ConnectionString{
		ConnectionString: connectionString,
	}
}

func (c ConnectionString) GetConnectionString(...ConnectionStringOpt) string {
	return c.ConnectionString
}
