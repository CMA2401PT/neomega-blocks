# PyMCTranslate 转换数据来源

本项目的离线脚本使用 [Amulet-Team/PyMCTranslate](https://github.com/Amulet-Team/PyMCTranslate) 的默认属性和跨版本映射，生成运行时的默认属性表、转入表及 Java 输出表。生成结果还经过 NEMC palette 和本项目转换规则的整理。

生成环境通过根目录的 `pyproject.toml` 和 `uv.lock` 固定上游 [1.2.30](https://github.com/Amulet-Team/PyMCTranslate/tree/1.2.30/PyMCTranslate)，其 build number 为 `317`，Java 输出目标为 1.21.4。上游库安装在被 Git 忽略的 `.venv/` 中。NEMC 的工作台、熔炉和点燃熔炉具有额外的 `custom_appearance` 属性，由 [pymc_specification.py](../gen_resource/block-converts/pymc_specification.py) 在生成时补充；该规则不修改上游安装文件。

Required Notice: Copyright Amulet Team. (https://www.amuletmc.com/)

该上游版本使用 **Amulet Team License 1.0.0**。原文保存在 [LICENSE](LICENSE)，来源为 [1.2.30/LICENSE](https://github.com/Amulet-Team/PyMCTranslate/blob/1.2.30/LICENSE)。它包含非商业、教育和非竞争等条款，不应视为 MIT 或无条件的商业使用授权。分发相关代码和数据时，应同时保留条款与 Required Notice。更换上游版本时，需要核对该版本自身的许可。
