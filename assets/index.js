function setSignupVal(){
    let f =document.getElementById("contactForm");
    let firstName=f.firstName.value;
    let lastName= f.lastName.value;

}





$().ready( function() {
    console.log('ready state');
    $('#save').on("click",function(){
        console.log('button log');
    let jsonData=$("#contactForm").serialize()
        console.log(jsonData);
        setSignupVal();
        $.post( "/formData", jsonData)
            .done(function( data ) {
                console.log( "Data Loaded: " + (JSON.parse(data)).data );
            });
        let firstName = $("#firstName").val(),
         lastName = $("#lastName").val();
        let phoneNumber = $("#phone").val();
        let fullName =firstName +' '+ lastName;
      /* $(".accordion").append('<h3 id="header">'+fullName+'</h3>'+'<div class="contactInfo" '+ '' +
           '<p id="item">'+phoneNumber+'</p'+
           '</div>'
       ); */
       /* $( function() {
            $( ".accordion").accordion();
        } ); */
        $(".group-list").append('<button class="accordion">'+fullName+'</button>'+'<div class="panel" '+ '' +
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
        }
    })

});

