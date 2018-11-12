jQuery(document).ready(function () {
    jQuery("#logoutBtn").click(function () {
        if (jQuery("#" + jQuery(this).data("fid")).length > 0) {
            jQuery("#" + jQuery(this).data("fid")).submit();
        }
    });

    jQuery('.summernote').summernote();
});