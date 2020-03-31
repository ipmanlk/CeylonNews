const showCoronaDialog = () => {
    const dialog = document.getElementById('corona-dialog');

    $.getJSON("https://hpb.health.gov.lk/api/get-current-statistical", (data) => {
        $("#corona-new-cases").html(data.data.local_new_cases);
        $("#corona-in-hospital").html(data.data.local_total_number_of_individuals_in_hospitals);
        $("#corona-recovered").html(data.data.local_recovered);
        $("#corona-deaths").html(data.data.local_deaths);
        $("#corona-total-cases").html(data.data.local_total_cases);

        dialog.show();
        
    }).fail((e) => {
        ons.notification.alert("Sorry!. Something went wrong.");
        console.log(e);
    });
};

const hideCoronaDialog = () => {
    const dialog = document.getElementById('corona-dialog');
    dialog.hide();
};