"""Verrou read-only : telegram_service ne doit jamais appeler d'écriture."""

import pathlib

FORBIDDEN = [
    "send_message",
    "send_file",
    "send_",
    "join",
    "JoinChannel",
    "LeaveChannel",
    "edit_message",
    "delete_messages",
    "SendReaction",
    "mark_read",
    "markRead",
    "create_channel",
    "invite",
    "kick",
    "ban",
]

# send_code_request / sign_in sont les seules écritures tolérées (login initial).
ALLOWED_CONTEXT = {"send_code_request", "sign_in"}


def test_no_write_calls():
    src = (
        pathlib.Path(__file__).parent.parent / "app" / "telegram_service.py"
    ).read_text()
    code_lines = []
    for line in src.splitlines():
        s = line.strip()
        if s.startswith("#") or s.startswith('"""') or s.startswith("INTERDIT"):
            continue
        if "DIT ici" in line or "ou tout appel" in line:
            continue
        if "send_code_request" in line or "sign_in" in line:
            continue  # login initial uniquement
        code_lines.append(line)
    code = "\n".join(code_lines)
    for token in FORBIDDEN:
        if token == "send_":
            assert "client.send_" not in code and ".send_file" not in code, (
                "appel d'écriture interdit détecté"
            )
            continue
        if token == "join":
            assert "JoinChannel" not in code and "LeaveChannel" not in code
            assert "client.join" not in code
            continue
        assert token not in code, f"token interdit trouvé: {token}"
