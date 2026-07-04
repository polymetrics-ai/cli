package conformance
import ("testing";"polymetrics.ai/internal/connectors/defs";"polymetrics.ai/internal/connectors/engine")
func TestDumpCT(t *testing.T){
 bundles,_:=engine.LoadAll(defs.FS)
 for _,b:=range bundles{ if b.Name!="commercetools"{continue}
  rep:=RunBundle(b)
  for _,c:=range rep.Checks{ t.Logf("check=%s passed=%v skipped=%v err=%s",c.Name,c.Passed,c.Skipped,c.Error) }
 }
}
