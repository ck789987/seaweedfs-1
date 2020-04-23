package meta_cache

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/chrislusf/seaweedfs/weed/filer2"
	"github.com/chrislusf/seaweedfs/weed/glog"
	"github.com/chrislusf/seaweedfs/weed/pb/filer_pb"
	"github.com/chrislusf/seaweedfs/weed/util"
)

func SubscribeMetaEvents(mc *MetaCache, client filer_pb.FilerClient, dir string) error {

	processEventFn := func(resp *filer_pb.SubscribeMetadataResponse) error {
		message := resp.EventNotification
		ctx := context.Background()
		var err error
		if message.OldEntry != nil {
			key := util.NewFullPath(resp.Directory, message.OldEntry.Name)
			glog.V(4).Infof("deleting %v", key)
			err = mc.DeleteEntry(ctx, key)
		}

		if message.NewEntry != nil {
			dir := resp.Directory
			if message.NewParentPath != "" {
				dir = message.NewParentPath
			}
			key := util.NewFullPath(dir, message.NewEntry.Name)
			glog.V(4).Infof("creating %v", key)
			err = mc.InsertEntry(ctx, filer2.FromPbEntry(dir, message.NewEntry))
		}
		return err
	}

	var lastTsNs int64
	for {
		err := client.WithFilerClient(func(client filer_pb.SeaweedFilerClient) error {
			stream, err := client.SubscribeMetadata(context.Background(), &filer_pb.SubscribeMetadataRequest{
				ClientName: "mount",
				PathPrefix: dir,
				SinceNs:    lastTsNs,
			})
			if err != nil {
				return fmt.Errorf("subscribe: %v", err)
			}

			for {
				resp, listenErr := stream.Recv()
				if listenErr == io.EOF {
					return nil
				}
				if listenErr != nil {
					return listenErr
				}

				if err := processEventFn(resp); err != nil {
					return fmt.Errorf("process %v: %v", resp, err)
				}
				lastTsNs = resp.TsNs
			}
		})
		if err != nil {
			glog.V(0).Infof("subscribing filer meta change: %v", err)
			time.Sleep(time.Second)
		}
	}
}