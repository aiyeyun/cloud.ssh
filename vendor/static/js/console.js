(function (factory) {
    "use strict";
    if (typeof define === "function" && (define.amd || define.cmd)) {
        define(["jquery"], factory);
    } else {
        factory((typeof(jQuery) != "undefined") ? jQuery : window.Zepto);
    }
}
(function ($) {
    "use strict";
    $.fn.OpenTerminal = function (options) {
        if (options === undefined) {
            options = {};
        }
        var wsaddr, $console = this;
        wsaddr = options.wsaddr === undefined ? "ws://127.0.0.1:10086/ssh" : options.wsaddr;

        //调整终端大小
        var resizeTerminal = function(t, c, r) {
            var nav_height = $("nav").height();
            var body_height = $(window).height();
            var terminal_height = body_height - nav_height;
            $("#terminal").height(terminal_height);
            t.resize(c, r);
            $(".xterm-viewport").height(terminal_height)
            $(".xterm-scroll-area").height(terminal_height)
        };

        var getSize = function () {

            function getCols() {
                var div = "<div style='display: inline;display: none' id='test_h'>&nbsp;</div>"
                $console.append(div);
                var test_width = $("#test_h").width()
                $("#test_h").remove()
                return parseInt(($(window).width() / test_width) / 2.299) + 3
            }

            function getRows() {
                var div = "<div style='display: none' id='test'>&nbsp;</div>"
                $console.append(div);
                var test_height = $("#test").height()
                var nav_height = $("nav").height()
                var body_height = $(window).height()
                var size = parseInt(body_height - nav_height) / parseInt(test_height)
                size = parseInt(size)
                $("#test").remove();
                return size
            }

            return {
                cols: getCols(),
                rows: getRows()
            }
        }

        window.WebSocket = window.WebSocket || window.MozWebSocket;
        var cols = getSize().cols;
        var rows = getSize().rows;
        var term = null;
        var socket = new WebSocket(wsaddr + "&cols=" + cols + "&rows=" + rows);

        socket.onopen = function() {
            term = new Terminal({
                termName: "xterm",
                cols: cols,
                rows: rows,
                useStyle: true,
                convertEol: true,
                screenKeys: true,
                cursorBlink: false,
                visualBell: true,
                colors: Terminal.xtermColors
            });

            term.attach(socket);
            term._initialized = true;

            term.open($console.get(0));
            term.fit();

            resizeTerminal(term, cols, rows);

            $(window).resize(function() {
                resizeTerminal(term, getSize().cols, getSize().rows);
            });

            term.on("title", function(title) {
                $(document).prop("title", title);
            });

            window.term = term;
            window.socket = socket;

        };
        socket.onclose = function(e) {
            $('.modal').modal({
                    dismissible: true, // Modal can be dismissed by clicking outside of the modal
                    opacity: .5, // Opacity of modal background
                    in_duration: 300, // Transition in duration
                    out_duration: 200, // Transition out duration
                    starting_top: '4%', // Starting top style attribute
                    ending_top: '10%', // Ending top style attribute
                    ready: function(modal, trigger) { // Callback for Modal open. Modal and trigger parameters available.
                    },
                    complete: function() {
                        term.destroy();
                        window.location.href = "/logout"
                    } // Callback for Modal close
                }
            );
            $('#modal1').modal('open');
        };
        socket.onerror = function(e) {
            console.log("Socket error:", e);
        };

    }
}))