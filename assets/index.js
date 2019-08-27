
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

       // $(".group-list").append('<div class="accordion">'+'<h3>'+fullName+'</h3>'+'<div>' + '' + '<p>'+phoneNumber+'</p'+ '</div>' +'</div>');

     /*   $(".group-list").append('<button class="accordion">'+fullName+'</button>'+'<div class="panel" '+ '' +
            '<p>'+phoneNumber+'</p'+
            '</div>'
        );
        let acc = document.getElementsByClassName("accordion");
        let i;

        for (i = 0; i < acc.length; i++) {
            acc[i].addEventListener("click", function() {
                this.classList.toggle("active");
                let panel = this.nextElementSibling;
                if (panel.style.display === "block") {
                    panel.style.display = "none";
                } else {
                    panel.style.display = "block";
                }
            });
        } */
    })

});

