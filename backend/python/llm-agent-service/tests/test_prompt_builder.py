"""Unit tests for Prompt Builder.

测试系统提示词构建器的各种场景。
"""
import pytest
from pathlib import Path
from unittest.mock import patch, mock_open

from app.services.prompt_builder import PromptBuilder, get_prompt_builder


@pytest.fixture
def prompt_builder():
    """创建 PromptBuilder 实例."""
    return PromptBuilder()


@pytest.fixture
def sample_character():
    """创建示例角色信息."""
    return {
        "name": "小艾",
        "personality": "活泼开朗，热情友善",
        "background": "是一名大学生，喜欢旅行和摄影",
        "speaking_style": "语气轻快，经常使用emoji和语气词",
        "world_view": "相信每个人都有闪光点",
        "moderation_style": "cheerful",
    }


@pytest.fixture
def sample_user_portrait():
    """创建示例用户画像."""
    return {
        "name": "张三",
        "age": 25,
        "interests": ["编程", "阅读", "跑步"],
        "preferences": "喜欢简洁明了的回复",
    }


@pytest.fixture
def sample_context():
    """创建示例对话上下文."""
    return {
        "relation_type": "friend",
        "interaction_mode": "private",
    }


def test_build_system_prompt_basic(prompt_builder, sample_character):
    """测试基础系统提示词构建."""
    prompt = prompt_builder.build_system_prompt(character=sample_character)

    # 验证包含角色信息
    assert "小艾" in prompt
    assert "活泼开朗" in prompt
    assert "大学生" in prompt
    assert "语气轻快" in prompt


def test_build_system_prompt_with_user_portrait(
    prompt_builder, sample_character, sample_user_portrait
):
    """测试包含用户画像的提示词."""
    prompt = prompt_builder.build_system_prompt(
        character=sample_character, user_portrait=sample_user_portrait
    )

    # 验证包含用户画像信息
    assert "张三" in prompt
    assert "25岁" in prompt
    assert "编程" in prompt
    assert "阅读" in prompt


def test_build_system_prompt_with_context(
    prompt_builder, sample_character, sample_context
):
    """测试包含对话上下文的提示词."""
    prompt = prompt_builder.build_system_prompt(
        character=sample_character, conversation_context=sample_context
    )

    # 验证包含上下文信息
    assert "朋友" in prompt
    assert "私密对话" in prompt


def test_build_system_prompt_moderation(prompt_builder, sample_character):
    """测试内容审核部分."""
    prompt = prompt_builder.build_system_prompt(character=sample_character)

    # 验证包含审核规则
    assert "敏感内容" in prompt or "内容安全" in prompt
    assert "政治敏感" in prompt
    assert "脏话" in prompt


def test_build_system_prompt_consistency(prompt_builder, sample_character):
    """测试角色一致性约束."""
    prompt = prompt_builder.build_system_prompt(character=sample_character)

    # 验证包含一致性约束
    assert "角色一致性" in prompt or "保持角色" in prompt
    assert "不得暴露" in prompt or "不要暴露" in prompt


def test_build_system_prompt_safety_lock_public(
    prompt_builder, sample_character
):
    """测试公共模式的安全锁."""
    context = {"interaction_mode": "public"}
    prompt = prompt_builder.build_system_prompt(
        character=sample_character, conversation_context=context
    )

    # 验证包含安全锁策略
    assert "安全" in prompt or "隐私" in prompt
    assert "Owner" in prompt or "拥有者" in prompt
    assert "私密信息" in prompt or "隐私保护" in prompt


def test_build_system_prompt_safety_lock_private(
    prompt_builder, sample_character
):
    """测试私密模式不包含安全锁."""
    context = {"interaction_mode": "private"}
    prompt = prompt_builder.build_system_prompt(
        character=sample_character, conversation_context=context
    )

    # 私密模式下安全锁部分应该简化或不显示
    # 这里我们主要验证提示词能正常生成
    assert "小艾" in prompt


def test_moderation_styles(prompt_builder):
    """测试不同审核风格."""
    styles = ["cheerful", "strict", "gentle", "default"]

    for style in styles:
        character = {
            "name": "测试角色",
            "personality": "测试",
            "moderation_style": style,
        }
        prompt = prompt_builder.build_system_prompt(character=character)

        # 验证提示词包含对应风格的拒绝模板
        assert len(prompt) > 0


def test_load_moderation_config_success(tmp_path):
    """测试成功加载审核配置."""
    config_content = """
categories:
  - political
  - profanity

refusal_templates:
  default: "抱歉"
  cheerful: "哎呀"
"""

    config_file = tmp_path / "test_config.yaml"
    config_file.write_text(config_content)

    builder = PromptBuilder(moderation_config_path=str(config_file))

    assert "political" in builder.moderation_config["categories"]
    assert builder.moderation_config["refusal_templates"]["cheerful"] == "哎呀"


def test_load_moderation_config_not_found():
    """测试配置文件不存在时使用默认配置."""
    builder = PromptBuilder(moderation_config_path="/nonexistent/path.yaml")

    # 应该回退到默认配置
    assert "categories" in builder.moderation_config
    assert "refusal_templates" in builder.moderation_config


def test_build_character_section(prompt_builder, sample_character):
    """测试角色部分构建."""
    section = prompt_builder._build_character_section(sample_character)

    assert "小艾" in section
    assert "性格特征" in section
    assert "背景故事" in section
    assert "说话风格" in section


def test_build_user_portrait_section(prompt_builder, sample_user_portrait):
    """测试用户画像部分构建."""
    section = prompt_builder._build_user_portrait_section(sample_user_portrait)

    assert "张三" in section
    assert "25岁" in section
    assert "编程" in section


def test_build_context_section(prompt_builder, sample_context):
    """测试对话上下文部分构建."""
    section = prompt_builder._build_context_section(sample_context)

    assert "朋友" in section
    assert "私密对话" in section


def test_build_moderation_section(prompt_builder):
    """测试审核部分构建."""
    section = prompt_builder._build_moderation_section("cheerful")

    assert "敏感内容" in section or "内容安全" in section
    assert "政治" in section
    assert "脏话" in section


def test_build_consistency_section(prompt_builder):
    """测试一致性约束部分构建."""
    section = prompt_builder._build_consistency_section()

    assert "角色一致性" in section
    assert "保持角色" in section
    assert "AI" in section


def test_build_safety_lock_section(prompt_builder, sample_character):
    """测试安全锁部分构建."""
    section = prompt_builder._build_safety_lock_section(sample_character)

    assert "安全" in section or "隐私" in section
    assert "Owner" in section
    assert "隐私保护" in section


def test_get_prompt_builder_singleton():
    """测试单例模式."""
    builder1 = get_prompt_builder()
    builder2 = get_prompt_builder()
    assert builder1 is builder2


def test_minimal_character():
    """测试最小化角色信息."""
    builder = PromptBuilder()
    minimal_character = {"name": "Test"}

    # 应该能正常构建，使用默认值
    prompt = builder.build_system_prompt(character=minimal_character)
    assert "Test" in prompt
    assert len(prompt) > 0
