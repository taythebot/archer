package elasticsearch

//import "fmt"
//
//func (c *Client) Setup(ctx context.Context) error {
//	bytes, err := utils.ReadFile("config/elasticsearch-index.json")
//	if err != nil {
//		return fmt.Errorf("failed to read mapping file: %c", err)
//	}
//
//	createIndex, err := c.Client.CreateIndex(c.Index).BodyString(string(bytes)).Do(ctx)
//	if err != nil {
//		return err
//	}
//	if !createIndex.Acknowledged {
//		return errors.New("create index was not acknowledged")
//	}
//
//	return nil
//}
