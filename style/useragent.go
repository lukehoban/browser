// Package style provides a default user-agent stylesheet.
// CSS 2.1 §6.4.4: User agent style sheets
package style

import (
	"github.com/lukehoban/browser/css"
)

// DefaultUserAgentStylesheet returns a comprehensive user-agent stylesheet
// with default styles for common HTML elements.
// Based on CSS 2.1 Appendix D and modern browser defaults.
func DefaultUserAgentStylesheet() *css.Stylesheet {
	// Parse the default CSS rules
	// These provide sensible defaults matching common browser behavior
	defaultCSS := `
/* CSS 2.1 §17.2: Table default styles */
table { display: table; border-spacing: 2px; }
tr { display: table-row; }
td, th { display: table-cell; padding: 1px; }

/* CSS 2.1 §9.2.1: Block-level elements */
div, p, h1, h2, h3, h4, h5, h6, ul, ol, li, dl, dt, dd, 
blockquote, pre, form, fieldset, hr, address, center {
	display: block;
}

/* Heading margins and font sizes - HTML5 §10.3.1 */
h1 { font-size: 2em; margin: 0.67em 0; font-weight: bold; }
h2 { font-size: 1.5em; margin: 0.83em 0; font-weight: bold; }
h3 { font-size: 1.17em; margin: 1em 0; font-weight: bold; }
h4 { font-size: 1em; margin: 1.33em 0; font-weight: bold; }
h5 { font-size: 0.83em; margin: 1.67em 0; font-weight: bold; }
h6 { font-size: 0.67em; margin: 2.33em 0; font-weight: bold; }

/* Paragraph margins */
p { margin: 1em 0; }

/* List margins and padding */
ul, ol { margin: 1em 0; padding-left: 40px; }
li { display: list-item; }

/* Links - CSS 2.1 §16.3.1 */
a { color: #0000EE; text-decoration: underline; }
a:visited { color: #551A8B; }

/* Text formatting elements - HTML5 §10.3.1 */
b, strong { font-weight: bold; }
i, em { font-style: italic; }
u { text-decoration: underline; }
code, kbd, samp, tt { font-family: monospace; }
small { font-size: 0.83em; }
big { font-size: 1.17em; }

/* Preformatted text */
pre { font-family: monospace; white-space: pre; margin: 1em 0; }

/* Horizontal rule */
hr { border-top: 1px solid black; margin: 0.5em 0; }

/* Forms */
input, textarea, select, button { 
	font-size: 1em;
	font-family: inherit;
}

/* Quotations */
blockquote { margin: 1em 40px; }

/* Center element - deprecated but still used */
center { text-align: center; }
`

	stylesheet := css.Parse(defaultCSS)
	return stylesheet
}
