package main

const wrappedHTMLBase = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf8" />
<style>
html, body, div, span, applet, object, iframe,
h1, h2, h3, h4, h5, h6, p, blockquote, pre,
a, abbr, acronym, address, big, cite, code,
del, dfn, em, img, ins, kbd, q, s, samp,
small, strike, strong, sub, sup, tt, var,
b, u, i, center,
dl, dt, dd, ol, ul, li,
fieldset, form, label, legend,
table, caption, tbody, tfoot, thead, tr, th, td,
article, aside, canvas, details, embed,
figure, figcaption, footer, header, hgroup,
menu, nav, output, ruby, section, summary,
time, mark, audio, video {
        margin: 0;
        padding: 0;
        border: 0;
        font-size: 100%;
        font: inherit;
        vertical-align: baseline;
}
/* HTML5 display-role reset for older browsers */
article, aside, details, figcaption, figure,
footer, header, hgroup, menu, nav, section {
        display: block;
}
body {
        line-height: 1;
}
ol, ul {
        list-style: none;
}
blockquote, q {
        quotes: none;
}
blockquote:before, blockquote:after,
q:before, q:after {
        content: '';
        content: none;
}
table {
        border-collapse: collapse;
        border-spacing: 0;
}
</style>

</head>
<body>
<script>
    document.addEventListener("DOMContentLoaded", function() {
        document.removeEventListener("DOMContentLoaded", arguments.callee, false);
        var imgs = document.getElementsByTagName("img");
        var f = function() {
            var complete = true;
            for (var i = 0; i != imgs.length; i ++) {
                if (!imgs[i].complete) {
                    complete = false;
                    break;
                }
            }
            if (complete) {
                document.getElementById('ImgLoadedFlagACHHcLIkD3').style.display = "block";
            } else {
                window.setTimeout(f, 5);
            }
        };
        f();
        window.setTimeout(function() {
        	document.getElementById('ImgLoadedFlagACHHcLIkD3').style.display = "block";
        }, 3000);
    });
</script>
<div id="ImgLoadedFlagACHHcLIkD3" style="display:none;">test</div>
<div id="ACHHcLIkD3">
`
