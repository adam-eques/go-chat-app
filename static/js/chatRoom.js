const chatList = document.getElementById('chatList');

const chatRooms = [
    "ELect 200L",
    "Mech Engr 300l",
    "Computer Science 100l",
    "Food Tech 200l",
    "ELect 200L",
];


const token = localStorage.getItem('token');
for (const prop in chatRooms) {
    chatList.innerHTML += `
    <form action="joinRoom" method="post">
        <input type="hidden" name="token" value="${token}" />
        <button type="submit">${chatRooms[prop]}</button>
    </form>
    `;
}
