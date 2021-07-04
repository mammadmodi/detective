function stopWaitingModal() {
    const lockModal = $("#lock-modal");
    const loadingCircle = $("#loading-circle");
    const form = $("#my-form");
    // re-enable the form
    lockModal.css("display", "none");
    loadingCircle.css("display", "none");
    form.children("input").each(function () {
        $(this).attr("readonly", false);
    });
}

function ajax() {
    let url = document.getElementById('url').value
    let data = {};
    data["url"] = url
    $.ajax({
        type: "POST",
        url: "analyze-url",
        data: JSON.stringify(data),
        success: function (data) {
            $('#html-version').html(data.result.html_version)
            $('#page-title').html(data.result.page_title)
            $('#headings-count-h1').html(data.result.headings_count.h1)
            $('#headings-count-h2').html(data.result.headings_count.h2)
            $('#headings-count-h3').html(data.result.headings_count.h3)
            $('#headings-count-h4').html(data.result.headings_count.h4)
            $('#headings-count-h5').html(data.result.headings_count.h5)
            $('#headings-count-h6').html(data.result.headings_count.h6)
            $('#external-links').html(data.result.links_count.external)
            $('#internal-links').html(data.result.links_count.internal)
            $('#inaccessible-links').html(data.result.inaccessible_links_count)
            let hasLoginForm = data.result.has_login_form
            let hasLoginFormMsg = "No"
            if (hasLoginForm === true) {
                hasLoginFormMsg = "Yes"
            }
            $('#has-login').html(hasLoginFormMsg)
            $('#result_box').show();
            let alert = $('#alert')
            alert.show()
            alert.removeClass("alert-danger");
            alert.addClass("alert-success");
            $('#alert_message').html("Success! Result for url: " + url)
            $('#heading_result_table').show()
            $('#result_table').show()
            stopWaitingModal()
        },
        error: function (xhr, status, error) {
            $('#result_box').show()
            let alert = $('#alert')
            alert.show()
            alert.removeClass("alert-success");
            alert.addClass("alert-danger");
            $('#alert_message').html("Failed! " + xhr.responseJSON.error)
            $('#heading_result_table').hide()
            $('#result_table').hide()
            stopWaitingModal()
        },
        dataType: "json",
        contentType: "application/json"
    });
}

$(document).ready(function () {
    const lockModal = $("#lock-modal");
    const loadingCircle = $("#loading-circle");
    const form = $("#my-form");

    form.on('submit', function (e) {
        e.preventDefault(); //prevent form from submitting
        const url = $("input[name=url]").val();
        // lock down the form
        lockModal.css("display", "block");
        loadingCircle.css("display", "block");

        form.children("input").each(function () {
            $(this).attr("readonly", true);
        });
        ajax()
        setTimeout(stopWaitingModal, 30000);
    });
});
