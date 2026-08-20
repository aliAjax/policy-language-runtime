package transport

import "context"

type CompileFunc func(context.Context, string) (bool, error)

type GRPCServer struct{ Compiler CompileFunc }

func NewGRPC() *GRPCServer { return &GRPCServer{} }
func (g *GRPCServer) Compile(ctx context.Context, source string) (bool, error) {
	if g.Compiler != nil {
		result := make(chan struct {
			ok  bool
			err error
		}, 1)
		go func() {
			ok, err := g.Compiler(ctx, source)
			result <- struct {
				ok  bool
				err error
			}{ok: ok, err: err}
		}()
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case out := <-result:
			return out.ok, out.err
		}
	}
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return source != "", nil
	}
}
