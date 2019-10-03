
$().ready( function() {
    $( function() {
        $( "#contactList").accordion({
            collapsible: true
        });
    });

    console.log('ready state');
    $('#save').on("click",function(){
        console.log('button log');
        let jsonData=$("#contactForm").serialize()

        console.log(jsonData);

        $.post( "/formData", jsonData)
            .done(function( data ) {
                console.log( "Data Loaded: " + (JSON.parse(data)).data );
            });

        let firstName = $("#firstName").val();
        let lastName = $("#lastName").val();
        let phoneNumber = $("#phone").val();
        let fullName =firstName +' '+ lastName;


    })


});

function loadContact(ID) {
    console.log(ID);
    $.post("/editContact",{
        contactID: ID,
    }).done(function(r){
        $("#contactID").val( r.ID );
        $("#firstName").val( r.FirstName );
        $("#lastName").val( r.LastName );
        $("#phone").val( r.Phone );
        $("#officePhone").val( r.OfficePhone );
        $("#city").val( r.City );
        $("#state").val( r.State );
        $("#zip").val( r.Zip );



    }).fail(function (r) {
        console.log(r);
    });
}

function clearContact() {
    console.log('clearContact()');
    $("#contactID").val('');
    $("#firstName").val('');
    $("#lastName").val('');
    $("#phone").val('');
    $("#officePhone").val('');
    $("#city").val('');
    $("#state").val('');
    $("#zip").val('');

}


function deleteContact(ID) {
    console.log('deleteContact()')
    $.post("/deleteContact", {
        contactID: ID,
    }).done(function () {
        console.log('Contact Deleted');
        clearContact();
    }).fail(function () {
        console.log('Contact was not deleted');
    });

}










