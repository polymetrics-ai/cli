package app

// composeTransportRegistry keeps Open's production transport composition in a
// dedicated file. Its order and error path are intentionally unchanged from
// the former inline block.
func (a *App) composeTransportRegistry() error {
	transports, stage, err := newIssueLabelWarehouseMediatedTransport(a)
	if err != nil {
		return err
	}
	a.transports = transports
	a.transportStage = stage
	return nil
}
