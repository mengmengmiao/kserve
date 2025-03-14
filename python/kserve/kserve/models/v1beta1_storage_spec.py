# Copyright 2022 The KServe Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# coding: utf-8

"""
    KServe

    Python SDK for KServe  # noqa: E501

    The version of the OpenAPI document: v0.1
    Generated by: https://openapi-generator.tech
"""


import pprint
import re  # noqa: F401

import six

from kserve.configuration import Configuration


class V1beta1StorageSpec(object):
    """NOTE: This class is auto generated by OpenAPI Generator.
    Ref: https://openapi-generator.tech

    Do not edit the class manually.
    """

    """
    Attributes:
      openapi_types (dict): The key is attribute name
                            and the value is attribute type.
      attribute_map (dict): The key is attribute name
                            and the value is json key in definition.
    """
    openapi_types = {
        'key': 'str',
        'parameters': 'dict(str, str)',
        'path': 'str',
        'schema_path': 'str'
    }

    attribute_map = {
        'key': 'key',
        'parameters': 'parameters',
        'path': 'path',
        'schema_path': 'schemaPath'
    }

    def __init__(self, key=None, parameters=None, path=None, schema_path=None, local_vars_configuration=None):  # noqa: E501
        """V1beta1StorageSpec - a model defined in OpenAPI"""  # noqa: E501
        if local_vars_configuration is None:
            local_vars_configuration = Configuration()
        self.local_vars_configuration = local_vars_configuration

        self._key = None
        self._parameters = None
        self._path = None
        self._schema_path = None
        self.discriminator = None

        if key is not None:
            self.key = key
        if parameters is not None:
            self.parameters = parameters
        if path is not None:
            self.path = path
        if schema_path is not None:
            self.schema_path = schema_path

    @property
    def key(self):
        """Gets the key of this V1beta1StorageSpec.  # noqa: E501

        The Storage Key in the secret for this model.  # noqa: E501

        :return: The key of this V1beta1StorageSpec.  # noqa: E501
        :rtype: str
        """
        return self._key

    @key.setter
    def key(self, key):
        """Sets the key of this V1beta1StorageSpec.

        The Storage Key in the secret for this model.  # noqa: E501

        :param key: The key of this V1beta1StorageSpec.  # noqa: E501
        :type: str
        """

        self._key = key

    @property
    def parameters(self):
        """Gets the parameters of this V1beta1StorageSpec.  # noqa: E501

        Parameters to override the default storage credentials and config.  # noqa: E501

        :return: The parameters of this V1beta1StorageSpec.  # noqa: E501
        :rtype: dict(str, str)
        """
        return self._parameters

    @parameters.setter
    def parameters(self, parameters):
        """Sets the parameters of this V1beta1StorageSpec.

        Parameters to override the default storage credentials and config.  # noqa: E501

        :param parameters: The parameters of this V1beta1StorageSpec.  # noqa: E501
        :type: dict(str, str)
        """

        self._parameters = parameters

    @property
    def path(self):
        """Gets the path of this V1beta1StorageSpec.  # noqa: E501

        The path to the model object in the storage. It cannot co-exist with the storageURI.  # noqa: E501

        :return: The path of this V1beta1StorageSpec.  # noqa: E501
        :rtype: str
        """
        return self._path

    @path.setter
    def path(self, path):
        """Sets the path of this V1beta1StorageSpec.

        The path to the model object in the storage. It cannot co-exist with the storageURI.  # noqa: E501

        :param path: The path of this V1beta1StorageSpec.  # noqa: E501
        :type: str
        """

        self._path = path

    @property
    def schema_path(self):
        """Gets the schema_path of this V1beta1StorageSpec.  # noqa: E501

        The path to the model schema file in the storage.  # noqa: E501

        :return: The schema_path of this V1beta1StorageSpec.  # noqa: E501
        :rtype: str
        """
        return self._schema_path

    @schema_path.setter
    def schema_path(self, schema_path):
        """Sets the schema_path of this V1beta1StorageSpec.

        The path to the model schema file in the storage.  # noqa: E501

        :param schema_path: The schema_path of this V1beta1StorageSpec.  # noqa: E501
        :type: str
        """

        self._schema_path = schema_path

    def to_dict(self):
        """Returns the model properties as a dict"""
        result = {}

        for attr, _ in six.iteritems(self.openapi_types):
            value = getattr(self, attr)
            if isinstance(value, list):
                result[attr] = list(map(
                    lambda x: x.to_dict() if hasattr(x, "to_dict") else x,
                    value
                ))
            elif hasattr(value, "to_dict"):
                result[attr] = value.to_dict()
            elif isinstance(value, dict):
                result[attr] = dict(map(
                    lambda item: (item[0], item[1].to_dict())
                    if hasattr(item[1], "to_dict") else item,
                    value.items()
                ))
            else:
                result[attr] = value

        return result

    def to_str(self):
        """Returns the string representation of the model"""
        return pprint.pformat(self.to_dict())

    def __repr__(self):
        """For `print` and `pprint`"""
        return self.to_str()

    def __eq__(self, other):
        """Returns true if both objects are equal"""
        if not isinstance(other, V1beta1StorageSpec):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, V1beta1StorageSpec):
            return True

        return self.to_dict() != other.to_dict()
