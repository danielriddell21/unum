package diagram

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const pngScale = 2

type Browser struct {
	browser  *rod.Browser
	launcher *launcher.Launcher
	mu       sync.Mutex
}

func chromiumPath() (string, bool) {
	if bin := os.Getenv("UNUM_CHROMIUM_BIN"); bin != "" {
		if _, err := os.Stat(bin); err == nil {
			return bin, true
		}
	}
	return launcher.LookPath()
}

func BrowserAvailable() bool {
	_, ok := chromiumPath()
	return ok
}

func NewBrowser() (*Browser, error) {
	bin, ok := chromiumPath()
	if !ok {
		return nil, fmt.Errorf("no chromium found: install a browser or set UNUM_CHROMIUM_BIN")
	}
	l := launcher.New().Bin(bin).Headless(true).Leakless(false)
	if os.Geteuid() == 0 {
		l = l.Set("no-sandbox")
	}
	u, err := l.Launch()
	if err != nil {
		l.Cleanup()
		return nil, fmt.Errorf("launch chromium: %w", err)
	}
	b := rod.New().ControlURL(u)
	if err := b.Connect(); err != nil {
		l.Cleanup()
		return nil, fmt.Errorf("connect chromium: %w", err)
	}
	return &Browser{browser: b, launcher: l}, nil
}

func (b *Browser) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	_ = b.browser.Close()
	b.launcher.Cleanup()
}

func (b *Browser) page(html string) (*rod.Page, error) {
	page, err := b.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("new page: %w", err)
	}
	if err := page.SetDocumentContent(html); err != nil {
		_ = page.Close()
		return nil, fmt.Errorf("set content: %w", err)
	}
	return page, nil
}

func (b *Browser) SVGToPNG(svg []byte) ([]byte, error) {
	return b.rasterize(svg, pngScale, 0, 0)
}

func (b *Browser) SVGToPNGSized(svg []byte, maxW, maxH int) ([]byte, error) {
	return b.rasterize(svg, 0, maxW, maxH)
}

func (b *Browser) rasterize(svg []byte, scale float64, maxW, maxH int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	html := `<!doctype html><html><head><style>` +
		`html,body{margin:0;padding:0;overflow:hidden}::-webkit-scrollbar{display:none}` +
		`</style></head><body>` + string(svg) + `</body></html>`
	page, err := b.page(html)
	if err != nil {
		return nil, err
	}
	defer page.Close() //nolint:errcheck // page discarded with the browser

	el, err := page.Element("svg")
	if err != nil {
		return nil, fmt.Errorf("locate svg: %w", err)
	}

	// A d2 SVG carries only a viewBox (no width/height), so the browser would
	// stretch it to the viewport. Size the element from its intrinsic aspect —
	// either by a fixed scale, or fit within maxW×maxH so the vector is
	// rasterised crisply at the target resolution rather than downscaled later.
	res, err := el.Eval(`(scale, maxW, maxH) => {
		const vb = this.viewBox && this.viewBox.baseVal;
		const iw = vb && vb.width ? vb.width : this.getBBox().width;
		const ih = vb && vb.height ? vb.height : this.getBBox().height;
		let w, h;
		if (maxW > 0 && maxH > 0) {
			const s = Math.min(maxW / iw, maxH / ih);
			w = iw * s; h = ih * s;
		} else {
			w = iw * scale; h = ih * scale;
		}
		this.setAttribute('width', w);
		this.setAttribute('height', h);
		this.style.width = w + 'px';
		this.style.height = h + 'px';
		this.style.maxWidth = 'none';
		this.style.maxHeight = 'none';
		return { w: Math.max(1, Math.ceil(w)), h: Math.max(1, Math.ceil(h)) };
	}`, scale, maxW, maxH)
	if err != nil {
		return nil, fmt.Errorf("measure svg: %w", err)
	}
	w, h := res.Value.Get("w").Int(), res.Value.Get("h").Int()
	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{Width: w, Height: h}); err != nil {
		return nil, fmt.Errorf("set viewport: %w", err)
	}

	png, err := el.Screenshot(proto.PageCaptureScreenshotFormatPng, 0)
	if err != nil {
		return nil, fmt.Errorf("screenshot svg: %w", err)
	}
	return png, nil
}
