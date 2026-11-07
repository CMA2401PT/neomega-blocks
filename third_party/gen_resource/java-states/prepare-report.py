"""从 Mojang 官方服务端生成 Java 方块报告，需要该版本支持的 Java 运行时。"""

import argparse
import io
import json
from pathlib import Path
import shutil
import subprocess
import urllib.request
import zipfile


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--version",
        required=True,
        help="与 java/v1.JavaTargetVersion 一致",
    )
    parser.add_argument("--java", default="java", help="java 可执行文件")
    args = parser.parse_args()
    tmp = Path(__file__).resolve().parent / "tmp"
    tmp.mkdir(parents=True, exist_ok=True)

    # manifest 给出的地址按上游 SHA-1 定位到所需资源。
    with urllib.request.urlopen("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json") as response:
        manifest = json.load(response)
    entry = next(version for version in manifest["versions"] if version["id"] == args.version)
    with urllib.request.urlopen(entry["url"]) as response:
        metadata = response.read()
    (tmp / "server-version.json").write_bytes(metadata)
    with urllib.request.urlopen(json.loads(metadata)["downloads"]["server"]["url"]) as response:
        server = response.read()
    (tmp / "server.jar").write_bytes(server)
    with zipfile.ZipFile(io.BytesIO(server)) as bundle:
        nested = next(
            name for name in bundle.namelist()
            if name.startswith("META-INF/versions/") and name.endswith(".jar")
        )
        with zipfile.ZipFile(io.BytesIO(bundle.read(nested))) as game:
            (tmp / "version.json").write_bytes(game.read("version.json"))

    subprocess.run(
        [
            args.java,
            "-DbundlerMainClass=net.minecraft.data.Main",
            "-jar",
            "server.jar",
            "--reports",
            "--output",
            "generated",
        ],
        cwd=tmp,
        check=True,
    )
    shutil.copyfile(
        src=tmp / "generated/reports/blocks.json",
        dst=tmp / "blocks.json",
    )


if __name__ == "__main__":
    main()
