function setSignupVal(){
    let f=document.getElementById("")
}





$().ready( function() {
    console.log('ready state');
    $('#save').on("click",function(){
        console.log('button log');
    let jsonData=$("#contactForm").serialize()
        console.log(jsonData);

        $.post( "/formData", jsonData)
            .done(function( data ) {
                console.log( "Data Loaded: " + (JSON.parse(data)).data );
            });

    })

});

