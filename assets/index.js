
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
    $('#editButton').on("click",function(){
        console.log('loading contact');
            $.post("/editContact",{
                contactID: ID,
            }).done(function(r){
                $("#contactID").val( r.ID );
                $("#firstName").val( r.FirstName );
                $("#lastName").val( r.LastName );
                $("#phone").val( r.Phone );
                $("#city").val( r.City );
                $("#state").val( r.State );
                $("#zip").val( r.Zip );



            })

    })

});

