package connector

type SDKConnector struct {

}

func (H *SDKConnector) Exec(sql string) (*Data, error) {
	panic("implement me")
}
